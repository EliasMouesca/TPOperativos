# List of projects
PROJECTS = cpu kernel filesystem memoria

# Default target
all: $(PROJECTS)

# Compile each project
$(PROJECTS):
	@echo "Compiling $@..."
	@cd $@ && go build -o $@

.PHONY: all $(PROJECTS)

