#!/bin/bash
SESSION=servers

tmux -2 new-session -d -s $SESSION

tmux split-window -h
tmux split-window -h
tmux split-window -h
tmux split-window -h

tmux select-layout even-horizontal

tmux select-pane -t 1
tmux send-keys "ssh $NETID@sp19-cs425-g54-01.cs.illinois.edu"

tmux select-pane -t 2
tmux send-keys "ssh $NETID@sp19-cs425-g54-02.cs.illinois.edu"

tmux select-pane -t 3
tmux send-keys "ssh $NETID@sp19-cs425-g54-03.cs.illinois.edu"

tmux select-pane -t 4
tmux send-keys "ssh $NETID@sp19-cs425-g54-04.cs.illinois.edu"

tmux select-pane -t 5
tmux send-keys "ssh $NETID@sp19-cs425-g54-05.cs.illinois.edu"

tmux attach-session -t $SESSION
