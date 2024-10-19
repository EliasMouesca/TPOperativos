# List of projects
PROJECTS = cpu memoria filesystem kernel 
TERMINAL=st
SHELL=fish
CODE_TO_RUN=../demo_code/easy.dino
TMUX_SESSION=dinos_session

.NOTPARALLEL:

# Default target
all: $(PROJECTS)

# Compile each project
$(PROJECTS):
	@echo "Compiling $@..."
	@cd $@ && go build -o $@

.PHONY: all magic $(PROJECTS)

# Magia negra que seguro solamente funciona en mi compu
magic: all
	tmux new-session -d -s $(TMUX_SESSION)
	tmux split-window -h      # Split horizontally
	tmux split-window -v      # Split vertically in the first pane
	tmux split-window -v -t 0 # Split vertically in the second pane
	tmux send-keys -t $(TMUX_SESSION):0.0 "cd memoria && ./memoria" C-m
	tmux send-keys -t $(TMUX_SESSION):0.1 "cd cpu && ./cpu" C-m
	tmux send-keys -t $(TMUX_SESSION):0.2 "cd filesystem && ./filesystem" C-m
	tmux send-keys -t $(TMUX_SESSION):0.3 "cd kernel && ./kernel $(CODE_TO_RUN)" C-m
	# Attach to the session
	tmux attach-session -t $(TMUX_SESSION)

