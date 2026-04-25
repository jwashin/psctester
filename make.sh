#!/bin/bash

# --- 1. Preparation ---
# Ensure the nested web structure exists for webdev
mkdir -p web/web
# Move source files into the nested 'web/web' structure if they are in web/
mv web/index.html web/psctester.dart web/psctester.css web/web/ 2>/dev/null

# --- 2. Build the Frontend ---
echo "Building Frontend (Dart)..."
cd web
# Output the compiled files directly to the project-level build folder
webdev build --output web:../build
cd ..

# --- 3. Build the Backend ---
echo "Building Backend (Go)..."
# Build the binary and place it inside the build folder
go build -C backend -o ../build/psctester

# --- 4. Syncing Assets & Scripts ---
echo "Syncing CGI scripts and Assets..."
# Copy the entire cgi-bin directory into build/
cp -r cgi-bin build/
cp -r assets build/

# Copy icons/media directly into build/ so they are next to the HTML
#cp -r assets/* build/ 2>/dev/null

echo "------------------------------------------------"
echo "Build complete! Your environment is ready."
echo "Go to the build folder and run the server:"
echo "cd build && ./psctester"