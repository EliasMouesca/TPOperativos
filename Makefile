# List of projects
PROJECTS=cpu memoria filesystem kernel 
#TERMINAL=st
#SHELL=fish
CODE_TO_RUN=../pruebas/PLANI_PROC
TAMANIO=32

#.NOTPARALLEL:

# Default target
all: $(PROJECTS)

# Compile each project
$(PROJECTS):
	@echo "Compiling $@..."
	@cd $@ && go build -o $@

test1:


.PHONY: all magic $(PROJECTS)

#Magia negra que seguro solamente funciona en mi compu
# TMUX_SESSION=dinos_session
# magic:
# 	tmux new-session -d -s $(TMUX_SESSION)
# 	tmux source-file ~/.tmux.config
# 	tmux split-window -h      # Split horizontally
# 	tmux split-window -v      # Split vertically in the first pane
# 	tmux split-window -v -t 0 # Split vertically in the second pane
# 	tmux send-keys -t $(TMUX_SESSION):0.0 "cd memoria && go run ." C-m
# 	tmux send-keys -t $(TMUX_SESSION):0.1 "cd cpu && go run ." C-m
# 	tmux send-keys -t $(TMUX_SESSION):0.2 "cd filesystem && go run ." C-m
# 	tmux send-keys -t $(TMUX_SESSION):0.3 "sleep 1; cd kernel && go run . $(CODE_TO_RUN) $(TAMANIO)" C-m
# 	tmux attach-session -t $(TMUX_SESSION)

