// Package ckasim holds the files the binary carries with it: the images it builds,
// the tasks it runs, and the exam UI it serves.
package ckasim

import "embed"

//go:embed images tasks web
var Assets embed.FS
