#!/usr/bin/env bash

# ^ This is a shebang. It tells the shell to run this script with the /usr/bin/env command, which will locate the correct interpreter for the script. (Bash in this case)

# This script is the main installer script for the go-pot project.
# As with all scripts you run from the internet, you should read this script before running it to ensure it meets your security standards.
# I have done my best to annotate the script to make it clear what it does and how it does it.



# Set Script to fail if any command fails.
# This incluides any command in a pipeline.
# See https://gist.github.com/mohanpedala/1e2ff5661761d3abd0385e8223e16425?permalink_comment_id=3945021 for more information.
set -eo pipefail

##############
# Parse Args #
##############

# -h or --help flag for help
# -y or --yes flag for automatic yes to prompts
# -s or --service flag for automatic service installation

# Parse the arguments passed to the script
while [ "$#" -gt 0 ]; do
  case "$1" in
    -h|--help)
      echo "This is the help message"
      exit 0
      ;;
    -y|--yes)
      # Set the automatic yes flag
      AUTO_YES=1
      shift 1
      ;;
    *)
      echo "Unknown argument: $1"
      exit 1
      ;;
  esac
done

#############
# VARIABLES #
#############
REPO_URL="ryanolee/go-pot"


##############
# Utilities  #
##############
confirm_interactive() {
  
  # This function takes a message and a default value and asks the user to confirm the message.
  # If the user types 'y' or 'Y' the function returns 0, otherwise it returns 1.
  
  # Check if the AUTO_YES flag is set
  if [ -n "$AUTO_YES" ]
  then
    # If the AUTO_YES flag is set, return 1
    return 0
  fi
  
  message=$1

  # Print the message to the user
  echo -n "$message"

  # Read the user input (Keep on asking until the user inputs something)
  while read -r response
  do
    # If the user input is 'y' or 'Y' return 0
    if [ "$response" = "y" ] || [ "$response" = "Y" ]
    then
      return 0
    # If the user input is 'n' or 'N' return 1
    elif [ "$response" = "n" ] || [ "$response" = "N" ]
    then
      echo "Exiting..."
      exit 1
    fi
    # If the user input is not 'y', 'Y', 'n', or 'N' ask the user to input again
    echo -n "Please enter 'y' or 'n': "
  done

}

############################
#  Download pre conditions #
############################

# This function checks if go-pot is already installed and asks the user if they want to overwrite it
# with a newer version if it is already installed. (or skip the installation if it is already at the latest version)
check_existing_go_pot() {
  # Check if go-pot is already installed
  if command -v go-pot &> /dev/null
  then
    GO_POT_VERSION=$(go-pot version --short)
    if [ "$GO_POT_VERSION" = "$LATEST_RELEASE" ]
    then
      echo "go-pot version $GO_POT_VERSION is already installed at latest version!"
      exit 0
    fi
    confirm_interactive  "go-pot is installed at version $GO_POT_VERSION. Would you like to update to $LATEST_RELEASE (y/n): "
  fi
}

# This function checks if the script is running as root
# in the event it is not it asks the user if they would like to rerun the script as root
assert_running_as_root() {
  # Check if the script is running as root
  if [ "$(id -u)" -ne 0 ]
  then
    echo "This script must be run as root in order to install go-pot correctly."
    echo "If you would prefer not to run this script as root, you can install go-pot manually by downloading the latest release from https://github.com/${REPO_URL}/releases/download"
    echo "Or use the docker image: docker run --rm -it ryanolee/go-pot:latest [COMMAND]"
    confirm_interactive "Would you like to rerun this script as root? (y/N): "

    # Rerun the script as root if the user inputs 'y' or 'Y'
    echo "Rerunning script as root... (You may be prompted for your password)"
    sudo bash "$0"
    exit 0
  fi
}



# Check for if required commands are installed
# Required commands are: curl, tar

# Tries to fin a suitable package manager and installs the required packages.
# tar: Required for extracting the downloaded zip file from the GitHub release.
# curl: Required for downloading the GitHub release zip file.

find_package_manager_and_install() {
  # Debian/Ubuntu based systems
  if command -v apt-get &> /dev/null
  then
    #Update package list
    echo "Updating package list for apt-get..."
    apt-get update

    #Install curl and tar
    apt-get install -y curl tar
  elif command -v apt &> /dev/null
  then
    #Update package list
    echo "Updating package list for apt..."
    apt update

    #Install curl and tar
    apt install -y curl tar
  # Fedora based systems (Preferred over yum)
  elif command -v dnf &> /dev/null
  then
    # Update package list
    echo "Updating package list for dnf..."
    dnf check-update || true # DNF returns 100 if there are updates that can be made, so we ignore the return code

    # Install curl and tar
    dnf install -y --skip-broken curl tar gzip
  # Red Hat based systems
  elif command -v yum &> /dev/null
  then
    # Update package list
    echo "Updating package list for yum..."
    yum check-update

    # Install curl and tar
    yum install -y curl tar
  elif command -v pacman &> /dev/null
  then
    # Arch based systems
    # Ask for user to install manually given this is a potentially more dangerous operation
    echo "Looks like you are using an Arch based system."
    echo "In order not to break things please install 'curl' and 'tar' manually."
    echo "If you are sure you want to continue, please uncomment the following lines. (Then rerun this script)"
    echo "pacman -Syu"
    echo "pacman -S curl tar"
# OpenSUSE based systems
  elif command -v zypper &> /dev/null
  then
    # Update package list
    echo "Updating package list for zypper..."
    zypper refresh

    # Install curl and tar
    zypper install -y curl tar gzip
# Alpine based systems
  elif command -v apk &> /dev/null
  then
    # Update package list
    echo "Updating package list for apk..."
    apk update 

    # Install curl and tar
    apk add curl tar
  # MacOS
  elif command -v brew &> /dev/null
  then
    # Update package list
    echo "Updating package list for brew..."
    brew update
    
    # Install curl and tar
    brew install curl tar
  elif command -v pkg &> /dev/null
  then
  # FreeBSD
    # Update package list
    echo "Updating package list for pkg..."
    pkg update

    # Install curl and tar
    pkg install -y curl tar
  elif command -v emerge &> /dev/null
  then
    # Gentoo

    # Install curl and tar
    emerge curl tar 
  else
    echo "No package manager found. Tried apt-get, apt, yum, dnf, pacman, zypper, apk, brew, pkg, emerge. Please install curl and tar manually or through your package manager of choice."
  fi
}

install_dependencies_if_missing() {
  # Check if curl is installed
  if ! command -v curl &> /dev/null || ! command -v tar &> /dev/null
  then
    confirm_interactive "curl or tar is not installed. Would you like to install them? (y/N): "

    echo "curl or tar is not installed. Installing..."
    find_package_manager_and_install
  fi
}

###################
# Download go-pot #
###################

download_go_pot_to_tmp() {
  echo "Downloading go-pot from $DOWNLOAD_URL"

  # Download the latest release of go-pot
  curl -L -o /tmp/go-pot.tar.gz $DOWNLOAD_URL

  # Extract the downloaded file
  tar -xzf /tmp/go-pot.tar.gz -C /tmp
}

move_go_pot_to_bin() {
  # Move the go-pot binary to /usr/local/bin
  echo "Moving go-pot to /usr/local/bin"
  mv /tmp/go-pot /usr/local/bin/go-pot
}


###############
# Main Script #
###############

# This is the main script that runs the installation process for go-pot.
# It takes the following steps:
# 1. Checks the script is running a root and attempts to elevate permissions if not.
# 2. Check for required dependencies and installed and will try to install if they are not present (curl, tar)
# 3. Determine the system architecture and platform to download the correct release of go-pot.
# 4. Fetch the latest release of go-pot from the GitHub repository and check it against the current version if it is already installed.
# 5. Download the latest release of go-pot to /tmp and extract it.
# 6. Move the go-pot binary to /usr/local/bin.
# 7. Print the version of go-pot installed.

echo "Checking for required dependencies..."

# Call the function to install dependencies if they are missing
assert_running_as_root  
install_dependencies_if_missing

# Determine system architecture
ARCH=$(uname -m)
# Convert to "amd64", "arm64", "368" or error out
if [ "$ARCH" = "x86_64" ]
then
  ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]
then
  ARCH="arm64"
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

# Determine system platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Determine the latest release version of the repository
LATEST_RELEASE=$(curl -s https://api.github.com/repos/${REPO_URL}/releases/latest | grep "tag_name" | cut -d '"' -f 4 | cut -c 2-)

# Determine the download URL for the latest release of go-pot
DOWNLOAD_URL=$(echo "https://github.com/${REPO_URL}/releases/download/v${LATEST_RELEASE}/go-pot_${LATEST_RELEASE}_${OS}_${ARCH}.tar.gz")

# Check if go-pot is already installed and ask the user if they want to overwrite it
check_existing_go_pot

# Download the latest release of go-pot to /tmp
download_go_pot_to_tmp
move_go_pot_to_bin

go-pot version


