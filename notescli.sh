#!/bin/bash

# this is for easily sending structured messages to the websocket server

NOTE_ID=$1
JWT=$2
TYPE=$3
PAYLOAD=$4

if [[ -z "$NOTE_ID" || -z "$JWT" || -z "$TYPE" ]]; then
  echo "Usage: $0 <note_id> <jwt> <type> <payload>"
  echo "Examples:"
  echo "  ./notescli.sh <note_id> <jwt> edit 'Hello world!'"
  echo "  ./notescli.sh <note_id> <jwt> typing '{\"is_typing\":true}'"
  echo "  ./notescli.sh <note_id> <jwt> cursor '{\"position\":128}'"
  exit 1
fi

# generates random content if requested
if [[ "$CONTENT" == "random" ]]; then
  RANDOM_MESSAGES=(
    "Collaborating live now..."
    "Adding new paragraph..."
    "Fixing typo in line 2"
    "User joined the room..."
  )
  
  RANDOM_INDEX=$(( RANDOM % ${#RANDOM_MESSAGES[@]} ))
  CONTENT="${RANDOM_MESSAGES[$RANDOM_INDEX]}"
fi

# ensures content is not empty
if [[ -z "$CONTENT" ]]; then
  echo "Error: content must be provided or set to 'random'"
  exit 1
fi

JSON_PAYLOAD=$(jq -nc \
  --arg t "$TYPE" \
  --arg c "$CONTENT" \
  '{type: $t, content: $c}')

echo "Sending message to note: $NOTE_ID"
echo "Payload: $JSON_PAYLOAD"

# sends the message to the websocket server
(echo "$JSON_PAYLOAD"; sleep 5 ) | websocat "ws://localhost:3000/ws/notes/$NOTE_ID?token=$JWT"
