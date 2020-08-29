PROJECT=health_check
VIRTUALENV_DIR=/var/virtualenvs/
PROJECT_VIRTUALENV_DIR=$(VIRTUALENV_DIR)$(PROJECT)/
ENV_ACTIVATE_SCRIPT=$(PROJECT_VIRTUALENV_DIR)bin/activate

boot: install-virtualenv install-deps
.PHONY: boot

install-virtualenv:
        sudo pip install --no-index --find-links wheelhouse virtualenv
        sudo mkdir -p $(PROJECT_VIRTUALENV_DIR)
        sudo virtualenv --extra-search-dir=wheelhouse $(PROJECT_VIRTUALENV_DIR)
.PHONY: install-virtualenv

install-deps:
        . $(ENV_ACTIVATE_SCRIPT); \
        [ -n "$(VIRTUAL_ENV)" ] || (echo "out of virtualenv"; exit 1); \
        pip install bs4 requests pyyaml html5lib
.PHONY: install-deps
~