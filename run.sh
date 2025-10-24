#!/bin/bash

echo "Starting Leetbot..."
if [ ! -f ".env" ]; then
    echo ".env file not found! Please create it from .env.example"
    echo "Run: cp .env.example .env"
    exit 1
fi
TOKEN=$(grep "DISCORD_TOKEN=" .env | cut -d'=' -f2)
if [ "$TOKEN" = "YOUR_ACTUAL_BOT_TOKEN_HERE" ] || [ -z "$TOKEN" ]; then
    echo "Please update DISCORD_TOKEN in .env with your actual Discord bot token!"
    echo ""
    echo "Discord Bot Setup:"
    echo "1. Go to https://discord.com/developers/applications"
    echo "2. Create/select your bot application"
    echo "3. Go to Bot section and copy the token"
    echo "4. Update .env file: DISCORD_TOKEN=your_token_here"
    echo ""
    echo "Bot Invite URL Setup:"
    echo "1. Go to OAuth2 -> URL Generator"
    echo "2. Select scopes: bot, applications.commands"
    echo "3. Select permissions: Send Messages, Use Slash Commands"
    echo "4. Use the generated URL to add bot to your server"
    exit 1
fi
echo "Environment configured"
echo "Building bot..."
make build

if [ $? -eq 0 ]; then
    echo "Build successful"
    echo "Starting bot..."
    echo ""
    echo "Available Commands:"
    echo "   Text: !problems airbnb, !help"
    echo "   Slash: /problems"
    echo ""
    echo "   Press Ctrl+C to stop"
    echo ""
    ./bin/leetbot
else
    echo "Build failed!"
    exit 1
fi
