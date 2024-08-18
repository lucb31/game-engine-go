#!/bin/bash                                                                     

shopt -s globstar

# Setup bin directory
mkdir -p bin/assets

# Convert PNG assets
for file in assets/*.png
do
  if [[ ! -f "$file" ]]
  then
      continue
  fi
  noExtension="${file/.png/}"
  output="bin/$noExtension.go"
  variable="${noExtension/assets\//}"
  # Convert to PascalCase
  variablePascal=$(echo "$variable" | sed -r 's/(^|_)(.)/\U\2/g')

  echo "go run github.com/hajimehoshi/file2byteslice/cmd/file2byteslice@latest -input $file -output $output -package assets -var $variablePascal"
  go run github.com/hajimehoshi/file2byteslice/cmd/file2byteslice@latest -input $file -output $output -package assets -var $variablePascal
done

# Convert CSV assets
for file in assets/*.csv
do
  if [[ ! -f "$file" ]]
  then
      continue
  fi
  noExtension="${file/.csv/_CSV}"
  output="bin/$noExtension.go"
  variable="${noExtension/assets\//}"
  # Convert to PascalCase
  variablePascal=$(echo "$variable" | sed -r 's/(^|_)(.)/\U\2/g')

  echo "go run github.com/hajimehoshi/file2byteslice/cmd/file2byteslice@latest -input $file -output $output -package assets -var $variablePascal"
  go run github.com/hajimehoshi/file2byteslice/cmd/file2byteslice@latest -input $file -output $output -package assets -var $variablePascal
done
