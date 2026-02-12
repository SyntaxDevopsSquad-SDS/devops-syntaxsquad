#!/bin/bash
VENV_PATH="backend/venv/bin/python3"
PYTHON_SCRIPT_PATH="backend/app.py"

while true; do
    "$VENV_PATH" "$PYTHON_SCRIPT_PATH"
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo "Script crashed with exit code $exit_code. Restarting..." >&2
        sleep 1
    else
        echo "Script exited cleanly. Stopping..." >&2
        break
    fi
done
