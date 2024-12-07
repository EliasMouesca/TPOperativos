# List of projects
PROJECTS=cpu memoria filesystem kernel 
#TERMINAL=st
#SHELL=fish
CODE_TO_RUN=../pruebas/PLANI_PROC

#.NOTPARALLEL:

# Default target
all: $(PROJECTS)

# Compile each project
$(PROJECTS):
	@echo "Compiling $@..."
	@cd $@ && go build -o $@

test1:


.PHONY: all magic $(PROJECTS)

# Magia negra que seguro solamente funciona en mi compu
# TMUX_SESSION=dinos_session
# magic: all
# 	tmux new-session -d -s $(TMUX_SESSION)
# 	tmux source-file ~/.tmux.config
# 	tmux split-window -h      # Split horizontally
# 	tmux split-window -v      # Split vertically in the first pane
# 	tmux split-window -v -t 0 # Split vertically in the second pane
# 	tmux send-keys -t $(TMUX_SESSION):0.0 "cd memoria && ./memoria" C-m
# 	tmux send-keys -t $(TMUX_SESSION):0.1 "cd cpu && ./cpu" C-m
# 	tmux send-keys -t $(TMUX_SESSION):0.2 "cd filesystem && ./filesystem" C-m
# 	tmux send-keys -t $(TMUX_SESSION):0.3 "sleep 1; cd kernel && ./kernel $(CODE_TO_RUN) 32" C-m
# 	tmux attach-session -t $(TMUX_SESSION)

