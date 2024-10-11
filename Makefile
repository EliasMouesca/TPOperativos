# List of projects
PROJECTS = cpu memoria kernel filesystem

# Default target
all: $(PROJECTS)

# Compile each project
$(PROJECTS):
	@echo "Compiling $@..."
	@cd $@ && go build -o $@

.PHONY: all $(PROJECTS)

