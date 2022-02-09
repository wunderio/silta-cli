#!/bin/bash

export CIRCLE_PROJECT_REPONAME=""
export CIRCLE_BRANCH="${1}"
export RELEASE_SUFFIX="${2}"

########################################################
## Slightly modified copy of orb code

# Make sure namespace is lowercase.
namespace="${CIRCLE_PROJECT_REPONAME,,}"

# # Create the namespace if it doesn't exist.
# if ! kubectl get namespace "$namespace" &>/dev/null ; then
#   kubectl create namespace "$namespace"
# fi

# # Tag the namespace if it isn't already tagged.
# if ! kubectl get namespace -l name=$namespace --no-headers | grep $namespace &>/dev/null ; then
#   kubectl label namespace "$namespace" "name=$namespace" --overwrite
# fi

# Make sure release name is lowercase without special characters.
branchname_lower="${CIRCLE_BRANCH,,}"
release_name="${branchname_lower//[^[:alnum:]]/-}"

suffix_test="${RELEASE_SUFFIX}"
declare -i total_length=${#suffix_test}+${#release_name}
suffix=''
if [[ -n "${RELEASE_SUFFIX}" && $total_length -gt 39 ]]; then
  suffix="${RELEASE_SUFFIX}"
  if [ ${#suffix} -gt 12 ]; then
    suffix="$(printf "$suffix" | cut -c 1-7)-$(printf "$suffix" | shasum -a 256 | cut -c 1-4 )"
  fi
  #Maximum length of a release name + release suffix. -1 is for separating '-' char before suffix
  declare -i rn_max_length=40-${#suffix}-1

  # Length of a shortened rn_max_length to allow for an appended hash
  declare -i rn_cut_length=rn_max_length-5

  # If name is too long, truncate it and append a hash
  if [ ${#release_name} -ge $rn_max_length ]; then
    release_name="$(printf "$release_name" | cut -c 1-${rn_cut_length})-$(printf "$branchname_lower" | shasum -a 256 | cut -c 1-4 )"
  fi
fi

silta_environment_name="${CIRCLE_BRANCH,,}"

if [[ -n "${RELEASE_SUFFIX}" ]]; then
  if [[ -n "$suffix" ]]; then
    # echo "Using suffix variable for release name"
    release_name="${release_name}-${suffix}"
    silta_environment_name="${CIRCLE_BRANCH,,}-${suffix}"
  else
    # echo "Using parameter for release name"
    release_name="${release_name}-${RELEASE_SUFFIX}"
    silta_environment_name="${CIRCLE_BRANCH,,}-${RELEASE_SUFFIX}"
  fi
fi


# echo "export RELEASE_NAME='$release_name'" >> "$BASH_ENV"
# echo "export NAMESPACE='$namespace'" >> "$BASH_ENV"
# echo "export SILTA_ENVIRONMENT_NAME='$silta_environment_name'" >> "$BASH_ENV"

# echo "The release name for this branch is \"$release_name\" in the \"$namespace\" namespace"
# echo "Release name: \"$release_name\""
echo "${release_name}"

# if helm status -n "$namespace" "$release_name" > /dev/null  2>&1
# then
#   current_chart_version=$(helm history -n "$namespace" "$release_name" --max 1 --output json | jq -r '.[].chart')
#   echo "export CURRENT_CHART_VERSION='$current_chart_version'" >> "$BASH_ENV"

#   echo "There is an existing chart with version $current_chart_version"
# fi