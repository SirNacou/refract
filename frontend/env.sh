#!/bin/sh

# Line endings must be \n, not \r\n
# Path to the generated file in your web server's static directory
OUTPUT_FILE=".output/public/config.js"

# Create directory if it doesn't exist
mkdir -p "$(dirname "$OUTPUT_FILE")"

# Start the file
echo "window._env_ = {" >$OUTPUT_FILE

# List variables you want to expose (or loop through all REACT_APP_ variables)
# Format: VAR_NAME: "VAR_VALUE",
echo "  API_URL: \"${VITE_API_URL}\"," >>$OUTPUT_FILE
echo "  DEFAULT_REDIRECT_URL: \"${VITE_DEFAULT_REDIRECT_URL}\"," >>$OUTPUT_FILE
echo "  ENVIRONMENT: \"${ENVIRONMENT}\"," >>$OUTPUT_FILE

# Close the object
echo "};" >>$OUTPUT_FILE

# Verify file was created
cat $OUTPUT_FILE
