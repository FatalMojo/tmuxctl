package main

import (
	"strconv"

	"github.com/alexandrebodin/tmuxctl/tmux"
)

type window struct {
	Sess   *session
	Name   string
	Dir    string
	Layout string
	Sync   bool
	Panes  []*pane
}

func newWindow(sess *session, config windowConfig) *window {
	win := &window{
		Sess:   sess,
		Name:   config.Name,
		Layout: config.Layout,
		Sync:   config.Sync,
	}

	if config.Dir != "" {
		win.Dir = lookupDir(config.Dir)
	} else {
		win.Dir = sess.Dir
	}

	if config.Layout == "" {
		win.Layout = "tiled"
	}

	for _, paneConfig := range config.Panes {
		win.Panes = append(win.Panes, newPane(win, paneConfig))
	}

	return win
}

func (w *window) start() error {
	args := []string{"new-window", "-t", w.Sess.Name, "-n", w.Name, "-c", w.Dir}
	_, err := tmux.Exec(args...)
	return err

}

func (w *window) init() error {
	var err error
	err = w.renderPane()
	if err != nil {
		return err
	}

	err = w.renderLayout()
	if err != nil {
		return err
	}

	err = w.zoomPanes()
	if err != nil {
		return err
	}

	if w.Sync {
		_, err := tmux.Exec("set-window-option", "-t", w.Sess.Name+":"+w.Name, "synchronize-panes")
		return err
	}

	return nil
}

func (w *window) renderPane() error {
	if len(w.Panes) == 0 {
		return nil
	}

	firstPane := w.Panes[0]
	if firstPane.Dir != "" && firstPane.Dir != w.Dir { // we need to move the pane
		_, err := tmux.Exec("send-keys", "-t", w.Sess.Name+":"+w.Name+"."+strconv.Itoa(w.Sess.TmuxOptions.PaneBaseIndex), "cd "+firstPane.Dir, "C-m")
		if err != nil {
			return err
		}
	}

	for _, pane := range w.Panes[1:] {
		args := []string{"split-window", "-t", w.Sess.Name + ":" + w.Name, "-c", pane.Dir}

		_, err := tmux.Exec(args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *window) renderLayout() error {
	_, err := tmux.Exec("select-layout", "-t", w.Sess.Name+":"+w.Name, w.Layout)
	return err
}

func (w *window) zoomPanes() error {
	for idx, pane := range w.Panes {
		if pane.Zoom {
			index := strconv.Itoa(idx + w.Sess.TmuxOptions.PaneBaseIndex)
			_, err := tmux.Exec("resize-pane", "-t", w.Sess.Name+":"+w.Name+"."+index, "-Z")
			if err != nil {
				return err
			}

			return nil // stop after first pane zoomed
		}
	}

	return nil
}
