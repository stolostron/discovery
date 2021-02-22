# Copyright Contributors to the Open Cluster Management project

#!/bin/bash

ALL_FILES=$(find . -name "*")

COPY_HEADER_FILE="$PWD/cicd-scripts/copyright-header.txt"

COPY_HEADER_STRING=$(cat $COPY_HEADER_FILE)
NEWLINES="\n\n"

for FILE in $ALL_FILES
do
    COMMENT_START="# "
    COMMENT_END=""
    if [[ $FILE  == *".go" ]]; then
        COMMENT_START="// "
    fi
    if [[ $FILE  == *".md" ]]; then
        COMMENT_START="[comment]: # ( "
        COMMENT_END=" )"
    fi

    if [[ $FILE  == *".go"       \
            || $FILE == *".yaml" \
            || $FILE == *".yml"  \
            || $FILE == *".sh"   \
            || $FILE == *"Dockerfile" \
            || $FILE == *"Makefile"  \
            || $FILE == *".gitignore"  \
            || $FILE == *".md"  ]]; then
        
        FILE_STRING=$(cat $FILE)
        HEADER_WITH_COMMENT="$COMMENT_START$COPY_HEADER_STRING$COMMENT_END"

        if [[ $FILE_STRING == "$HEADER_WITH_COMMENT"* ]]; then
            echo "$FILE : Header already exists! Skipping!"
        else
            echo -e "$COMMENT_START$COPY_HEADER_STRING$COMMENT_END$NEWLINES$FILE_STRING" > $FILE
            echo "$FILE : Adding copyright header to file!"
        fi
    else
        echo "$FILE : DO NOTHING!"
    fi
done
