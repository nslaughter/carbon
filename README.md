# carbon 

A tool for building text based projects

## About

The first version of carbon was hacked together in an evening to simplify creating project templates from existing work. It was a quick script for cases when building a template like `cookiecutter` would be overkill.

The goal is to eliminate the work of planning and defining project layouts and boilerplate with templates when they already exist in working code that can be used as reference. This is much simpler than building and maintaining templates to track lessons learned across projects.

## Workflow

Write your carbon script in a yaml file. By default the tool looks for `carbon.yaml` in the working directory. You can then script actions to get and change source.

## Design

carbon is designed as a pluggable framework. The logic for each behavior is contained in structs that implement the Action interface. Actions register by name and the script executor resolves names in the register.

