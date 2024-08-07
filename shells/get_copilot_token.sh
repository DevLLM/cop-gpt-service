#!/bin/bash

#========================================================
#   Description: Get Github Copilot token Tools
#   Usage: ./get_copilot_token.sh [option]
#========================================================

SCRIPT_VERSION="0.0.3"

red='\033[0;31m'
green='\033[0;32m'
yellow='\033[0;33m'
plain='\033[0m'

COPILOT_CLI_CONFIG_PATH=~/.copilot-cli-access-token
IDEA_COPILOT_PLUGIN_CONFIG_PATH=~/.config/github-copilot/hosts.json

echo -e "Github Copilot token Tools [v${SCRIPT_VERSION}]"

by_github_oauth() {
  if ! command -v python3 &> /dev/null
  then
      echo "python3 is not installed. Please install python3 first."
      exit
  fi
  if ! python3 -c "import requests" &> /dev/null
  then
      echo "requests is not installed. Installing..."
      pip3 install requests
  fi
  python3 <(curl -fsSL https://raw.githubusercontent.com/aaamoon/copilot-gpt4-service/master/shells/get_copilot_token.py)
}

by_idea_copilot_plugin() {
  if [ -f "${IDEA_COPILOT_PLUGIN_CONFIG_PATH}" ]; then
    echo -e "Your IDEA Github Copilot token: ${green}$(cat ${IDEA_COPILOT_PLUGIN_CONFIG_PATH})${plain}"
  else
    echo -e "${red}IDEA Github Copilot token not found. You need to install IDEA Github Copilot plugin first，and then authorize it. ${yellow}Plugins -> Marketplace -> Search Github Copilot -> Install${plain}."
  fi
  if [[ $# == 0 ]]; then
    before_show_menu
  fi
}

by_copilot_cli() {
  if ! command -v npm &> /dev/null
  then
      echo "npm is not installed. Please install npm first."
      exit
  fi

  if ! command -v github-copilot-cli &> /dev/null
  then
      echo "github-copilot-cli is not installed. Installing..."
      npm i @githubnext/github-copilot-cli -g
  fi

  # Check if the token file exists
  if [ -f "${COPILOT_CLI_CONFIG_PATH}" ]; then
      echo -e "Github Copilot CLI token found: ${green}$(cat ${COPILOT_CLI_CONFIG_PATH})${plain}"
      read -p "Do you want to reauthorize? (y/n) " -n 1 -r
      echo
      if [[ ! $REPLY =~ ^[Yy]$ ]]
      then
          exit 1
      fi
  else
      echo "Github Copilot CLI token not found. Please authorize."
  fi

  # Retrieve the token
  github-copilot-cli auth

  # Check if the token file exists
  if [ -f "${COPILOT_CLI_CONFIG_PATH}" ]; then
      echo -e "Your Github Copilot CLI token : ${green}$(cat ${COPILOT_CLI_CONFIG_PATH})${plain}"
  else
      echo -e "${red}Github Copilot CLI token not found. Please try again.${plain}"
  fi
  if [[ $# == 0 ]]; then
    before_show_menu
  fi
}

show_usage() {
  echo -e "\nUsage: ${green}get_copilot_token.sh [option]${plain}"
  echo "Options:"
  echo -e "  ${green}by_github_oauth${plain}: Get Github Copilot token by Github OAuth (Recommended)"
  echo -e "  ${green}by_copilot_cli${plain}: Get Github Copilot token by Github Copilot CLI"
  echo -e "  ${green}by_idea_copilot_plugin${plain}: Get Github Copilot token by IDEA Github Copilot plugin"
}

show_menu() {
  echo -e ">
    ${green}Get Github Copilot token Tools ${plain}[v${SCRIPT_VERSION}]${plain}
    ————————————————-
    ${green}1.${plain}  Get Github Copilot token by Github OAuth (Recommended)
    ${green}2.${plain}  Get Github Copilot token by IDEA Github Copilot plugin
    ${green}3.${plain}  Get Github Copilot token by Github Copilot CLI
    ————————————————-
    ${green}0.${plain}  Exit
    "
  echo && read -ep "Please enter a number [0-3]: " num

  case "${num}" in
  0)
    exit 0
    ;;
  1)
    by_github_oauth
    ;;
  2)
    by_idea_copilot_plugin
    ;;
  3)
    by_copilot_cli
    ;;
  *)
    echo -e "${red}Please enter the correct number [0-3]${plain}"
    ;;
  esac
}

before_show_menu() {
  echo && echo -n -e "${yellow}* Press Enter to return to the main menu. *${plain}" && read temp
  show_menu
}

if [ $# -gt 0 ] && [ -n "$1" ]; then
  case $1 in
  "by_copilot_cli")
    by_copilot_cli 0
    ;;
  "by_idea_copilot_plugin")
    by_idea_copilot_plugin 0
    ;;
  *)
    $@ || show_usage
    ;;
  esac
else
  show_menu
fi
