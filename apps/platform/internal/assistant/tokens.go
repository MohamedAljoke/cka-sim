package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/auth"
)

type AuthTokens struct {
	URL string
}

// Only auth signs tokens, so the user's session is traded there for one the assistant may hold.
func (a AuthTokens) Issue(ctx context.Context, r *http.Request) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.URL+"/api/auth/assistant-token", nil)
	if err != nil {
		return "", err
	}
	if cookie, err := r.Cookie(auth.CookieName); err == nil {
		req.AddCookie(cookie)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", domainerr.Unavailable("assistant_token", fmt.Sprintf("reach auth: %v", err)).AsFault()
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if res.StatusCode != http.StatusOK {
		detail := res.Status
		if p, ok := httpx.ParseProblem(body); ok {
			detail = p.Detail
		}
		return "", domainerr.Unavailable("assistant_token", "auth refused an assistant token: "+detail).AsFault()
	}
	var reply struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &reply); err != nil || reply.Token == "" {
		return "", domainerr.Unavailable("assistant_token", "auth sent no assistant token").AsFault()
	}
	return reply.Token, nil
}
