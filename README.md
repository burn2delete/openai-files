# openai-files
Sync local folder with OpenAI Vector Store
This Go script scans a specified folder for files and nested files, generates a manifest containing the path to the file and an SHA256 hash of the file, and updates this manifest when files change.

Additionally, the script:

Uploads files to the OpenAI Files endpoint.
Writes the File ID back to the manifest.
Synchronizes files with an OpenAI Vector Store.
Supports cleanup to remove files no longer found in the folder from both the OpenAI Files endpoint and the vector store.
Features
Folder Scanning: Recursively scans a given folder for files and generates a SHA256 hash for each file.
Manifest Generation: Creates or updates a manifest file that tracks file paths, hashes, and OpenAI File IDs.
File Uploading: Uploads new or changed files to the OpenAI Files endpoint.
Vector Store Synchronization: Adds or updates files in an OpenAI Vector Store.
Cleanup: Removes files from both the OpenAI Files endpoint and the vector store that are no longer present in the folder.
Dry-Run Mode: Option to disable uploading and deletion for testing purposes.
Logging: Includes operation metadata in the manifest for audit and logging purposes.
Usage
Prerequisites
Go language environment.
OpenAI API Key, set in your environment variables.
Environment Setup
Set the OpenAI API Key:

export OPENAI_API_KEY=your_openai_api_key_here
Prepare Your Directory:

Create a folder named your-folder and add some files to it.
Alternatively, specify another folder path with the --folder flag.
Running the Script
Normal Run
go run main.go --folder your-folder --vector-store-id <VECTOR_STORE_ID> --output manifest.json
Dry-Run Mode (Disables Uploading and Deletion)
go run main.go --dry-run --folder your-folder --vector-store-id <VECTOR_STORE_ID>
Cleanup Mode
go run main.go --cleanup --folder your-folder --vector-store-id <VECTOR_STORE_ID> --output manifest_updated.json
Command-Line Flags
--folder: Folder to scan for files (default: ./your-folder).
--vector-store-id: ID of the OpenAI Vector Store.
--cleanup: Enable cleanup of deleted files in OpenAI.
--dry-run: Disable uploading to OpenAI.
--output: Output file for the manifest; if not specified, print to console.
Example Commands
# Normal run with vector store and output file
go run main.go --folder your-folder --vector-store-id <VECTOR_STORE_ID> --output manifest.json

# Dry-run mode
go run main.go --dry-run --folder your-folder --vector-store-id <VECTOR_STORE_ID>

# Cleanup mode
go run main.go --cleanup --folder your-folder --vector-store-id <VECTOR_STORE_ID> --output manifest_updated.json
Script Overview
Main Function: Parses flags, reads the existing manifest, scans the folder for files, uploads files to OpenAI, and handles cleanup and synchronization with the vector store.
Hash Calculation: Generates SHA256 hash for each file.
File Uploading: Uploads files to the OpenAI Files endpoint.
Vector Store Management: Adds, updates, and removes files in the OpenAI Vector Store.
Manifest Handling: Creates or updates the manifest with file paths, hashes, and OpenAI File IDs.
Error Handling
The script includes basic error handling and logging for:

File access and read/write operations.
HTTP requests to the OpenAI API.
Feel free to report issues or request enhancements by opening an issue on the repository.
