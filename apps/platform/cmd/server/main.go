package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"
	"github.com/MohamedAljoke/cka-sim/apps/platform/catalog"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/assistant"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/auth"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/desk"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/exam"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm/vertex"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox/gke"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox/local"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/server"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

// Only this machine: the page's terminal can sudo on every node.
const addr = "127.0.0.1:7070"

type settings struct {
	AuthSecret string        `key:"CKA_SIM_AUTH_SECRET" minLength:"32"`
	AgentToken string        `key:"CKA_SIM_AGENT_TOKEN"`
	Provider   string        `key:"CKA_SIM_PROVIDER" default:""`
	LabURL     string        `key:"CKA_SIM_LAB_URL" default:"http://127.0.0.1:7171"`
	Pool       string        `key:"CKA_SIM_POOL" default:"labs"`
	LabIdle    time.Duration `key:"CKA_SIM_LAB_IDLE" default:"30m"`
	AuthURL    string        `key:"CKA_SIM_AUTH_URL" default:"http://127.0.0.1:7272"`
	MCPURL     string        `key:"CKA_SIM_MCP_URL" default:"http://host.docker.internal:7373/mcp"`
	LLM        string        `key:"CKA_SIM_LLM" default:""`
	Project    string        `key:"CKA_SIM_VERTEX_PROJECT" default:""`
	Location   string        `key:"CKA_SIM_VERTEX_LOCATION" default:"global"`
	Model      string        `key:"CKA_SIM_VERTEX_MODEL" default:"gemini-3.5-flash"`
	ChatMCPURL string        `key:"CKA_SIM_CHAT_MCP_URL" default:"http://127.0.0.1:7373/mcp"`
	ChatDaily  int           `key:"CKA_SIM_CHAT_DAILY" default:"100" min:"1"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var cfg settings
	if err := config.Load(&cfg); err != nil {
		return fmt.Errorf("%w (make dev sets these)", err)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("claim %s: %w", addr, err)
	}
	defer ln.Close()
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	chats, err := newChats(ctx, cfg)
	if err != nil {
		return err
	}
	all, err := tasks.Load(catalog.FS)
	if err != nil {
		return fmt.Errorf("load the task catalog: %w", err)
	}
	examDir, err := exam.DefaultDir()
	if err != nil {
		return err
	}
	desks := desk.New(openDesk(provider, all, examDir), time.Now)
	go desks.Watch(ctx, time.Minute, cfg.LabIdle)
	practice := server.Practice{
		Tasks:      all,
		Files:      catalog.FS,
		Exams:      true,
		Assistants: []assistant.Provider{assistant.Claude{}},
		Tokens:     assistant.AuthTokens{URL: cfg.AuthURL},
		MCPURL:     cfg.MCPURL,
		Chats:      chats,
		ChatTools:  assistant.MCPTools{URL: cfg.ChatMCPURL},
	}

	sessions := auth.Verifier{Secret: []byte(cfg.AuthSecret), Now: time.Now}
	srv := &http.Server{Handler: auth.Require(sessions, server.New(desks, practice))}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()
	fmt.Printf("backend on http://%s; Ctrl-C stops it\n", addr)
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Shutdown doesn't wait for websockets; ending the labs ends their shells.
	endCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return desks.EndAll(endCtx)
}

func newProvider(ctx context.Context, cfg settings) (sandbox.Provider, error) {
	switch cfg.Provider {
	case "":
		return &local.Provider{Agent: labagent.Client{URL: cfg.LabURL, Token: cfg.AgentToken}}, nil
	case "gke":
		p, err := gke.New(gke.Config{Pool: cfg.Pool, AgentToken: cfg.AgentToken})
		if err != nil {
			return nil, err
		}
		if err := p.ReapClaims(ctx); err != nil {
			return nil, fmt.Errorf("delete leftover lab claims: %w", err)
		}
		return p, nil
	default:
		return nil, fmt.Errorf("unknown CKA_SIM_PROVIDER %q; want gke or nothing", cfg.Provider)
	}
}

func newChats(ctx context.Context, cfg settings) ([]*assistant.Chat, error) {
	switch cfg.LLM {
	case "":
		return nil, nil
	case "vertex":
		if cfg.Project == "" {
			return nil, errors.New("CKA_SIM_LLM=vertex needs CKA_SIM_VERTEX_PROJECT")
		}
		model, err := vertex.New(ctx, vertex.Config{Project: cfg.Project, Location: cfg.Location, Model: cfg.Model})
		if err != nil {
			return nil, err
		}
		return []*assistant.Chat{{ID: "gemini", Model: model, MaxSteps: 8, MaxTokens: 2048, Daily: cfg.ChatDaily, Now: time.Now}}, nil
	default:
		return nil, fmt.Errorf("unknown CKA_SIM_LLM %q; want vertex or nothing", cfg.LLM)
	}
}

var accountID = regexp.MustCompile(`^[0-9A-Za-z-]+$`)

func openDesk(provider sandbox.Provider, all []tasks.Task, examDir string) desk.Opener {
	return func(id string) (*desk.Desk, error) {
		if !accountID.MatchString(id) {
			return nil, fmt.Errorf("account id %q is not safe for a file name", id)
		}
		lab := sandbox.NewLab(provider, id, time.Now)
		path := filepath.Join(examDir, id+".json")
		session, err := exam.Open(exam.Config{
			Store:   exam.FileStore{Path: path},
			Files:   catalog.FS,
			Runner:  lab,
			Catalog: all,
			Now:     time.Now,
			Rand:    rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
			Ready:   lab.Ensure,
		})
		if err != nil {
			return nil, fmt.Errorf("%w (delete %s to start fresh)", err, path)
		}
		return &desk.Desk{Lab: lab, Exam: session}, nil
	}
}
