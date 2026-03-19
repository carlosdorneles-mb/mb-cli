#!/bin/bash

# Replaces all occurrences of a substring in a string with another substring.
# Usage:
#   str_replace "input_string" "search_substring" "replace_substring"
# Example:
#   result=$(str_replace "hello world" "world" "universe")
#   echo "$result"  # Output: hello universe
str_replace() {
    local input="$1"
    local search="$2"
    local replace="$3"
    echo "${input//$search/$replace}"
}

# Converts a string to uppercase.
# Usage:
#   str_to_upper "your_string"
# Example:
#   result=$(str_to_upper "hello world")
#   echo "$result"  # Output: HELLO WORLD
str_to_upper() {
    local input
    input="$1"
    # Use tr for POSIX compatibility with shfmt
    printf '%s' "$input" | tr '[:lower:]' '[:upper:]'
}

# Converts a string to lowercase.
# Usage:
#   str_to_lower "YOUR_STRING"
# Example:
#   result=$(str_to_lower "HELLO WORLD")
#   echo "$result"  # Output: hello world
str_to_lower() {
    local input
    input="$1"
    # Use tr for POSIX compatibility with shfmt
    printf '%s' "$input" | tr '[:upper:]' '[:lower:]'
}

# Strips leading and trailing whitespace from a string.
# Usage:
#   str_trim "  your string  "
# Example:
#   result=$(str_trim "  hello world  ")
#   echo "$result"  # Output: hello world
str_trim() {
    echo "$1" | xargs
}

# Returns 0 if the string contains the given substring, 1 otherwise.
# Usage:
#   str_contains "your_string" "substring"
# Example:
#   if str_contains "hello world" "world"; then
#       echo "Contains substring"
#   fi
str_contains() {
    [[ "$1" == *"$2"* ]]
}

# Returns 0 if the string starts with the given prefix, 1 otherwise.
# Usage:
#   str_starts_with "your_string" "prefix"
# Example:
#   if str_starts_with "hello world" "hello"; then
#       echo "Starts with prefix"
#   fi
str_starts_with() {
    [[ "$1" == "$2"* ]]
}

# --- Array Helper Functions --- #

# Splits any comma-separated string elements in the referenced array into separate elements.
# Usage:
#   str_parse_comma_separated arr
# Example:
#   arr=("a,b,c") -> arr=("a" "b" "c")
str_parse_comma_separated() {
    local arr_name=$1
    local -a new_arr=()

    # Use indirect parameter expansion with (P) flag in zsh
    eval "local -a original_arr=(\"\${${arr_name}[@]}\")"

    for element in "${original_arr[@]}"; do
        if [[ "$element" == *","* ]]; then
            # Split on comma
            IFS=',' read -ra split_arr <<< "$element"
            new_arr+=("${split_arr[@]}")
        else
            new_arr+=("$element")
        fi
    done

    # Update the original array
    eval "${arr_name}=(\"\${new_arr[@]}\")"
}

# Joins all elements of the referenced array into a single comma-separated string element.
# Usage:
#   str_join_to_comma_separated arr
# Example:
#   arr=("a" "b" "c") -> arr=("a,b,c")
str_join_to_comma_separated() {
    local arr_name=$1
    local joined

    # Get array elements and join with comma
    eval "local -a arr_elements=(\"\${${arr_name}[@]}\")"

    # Join array elements with comma (POSIX-compatible)
    local first=1
    for element in "${arr_elements[@]}"; do
        if [ $first -eq 1 ]; then
            joined="$element"
            first=0
        else
            joined="${joined},${element}"
        fi
    done

    # Update the original array with single joined element
    eval "${arr_name}=(\"$joined\")"
}

# Converts a string to boolean. Returns 0 for true values (true, 1, on, yes), 1 for false.
# Usage:
#   str_to_bool "value"
# Example:
#   if str_to_bool "yes"; then
#       echo "true"
#   fi
str_to_bool() {
    local value
    value=$(str_to_lower "$1")
    case "$value" in
        "true" | "1" | "on" | "yes")
            return 0
            ;;
        "false" | "0" | "off" | "no")
            return 1
            ;;
        *)
            return 1
            ;;
    esac
}
