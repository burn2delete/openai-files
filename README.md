# openai-files
OpenAI Vector Store Sync

This Go script scans a specified folder for files and nested files, generates a manifest containing the path to the file and an SHA256 hash of the file, and updates this manifest when files change.

Additionally, the script:
- Uploads files to the OpenAI Files endpoint.
- Writes the File ID back to the manifest.
- Synchronizes files with an OpenAI Vector Store.
- Supports cleanup to remove files no longer found in the folder from both the OpenAI Files endpoint and the vector store.

## Features

- **Folder Scanning**: Recursively scans a given folder for files and generates a SHA256 hash for each file.
- **Manifest Generation**: Creates or updates a manifest file that tracks file paths, hashes, and OpenAI File IDs.
- **File Uploading**: Uploads new or changed files to the OpenAI Files endpoint.
- **Vector Store Synchronization**: Adds or updates files in an OpenAI Vector Store.
- **Cleanup**: Removes files from both the OpenAI Files endpoint and the vector store that are no longer present in the folder.
- **Dry-Run Mode**: Option to disable uploading and deletion for testing purposes.
- **Logging**: Includes operation metadata in the manifest for audit and logging purposes.

## Usage

### Prerequisites

- Go language environment.
- OpenAI API Key, set in your environment variables.

### Environment Setup

1. **Set the OpenAI API Key**:

    ```bash
    export OPENAI_API_KEY=your_openai_api_key_here
    ```

2. **Prepare Your Directory**:
    - Create a folder named `your-folder` and add some files to it.
    - Alternatively, specify another folder path with the `--folder` flag.

### Running the Script

#### Normal Run

```bash
go run main.go --folder your-folder --vector-store-id <VECTOR_STORE_ID> --output manifest.json
```

#### Dry-Run Mode (Disables Uploading and Deletion)
```bash
go run main.go --dry-run --folder your-folder --vector-store-id <VECTOR_STORE_ID>
```

#### Cleanup Mode

```bash
go run main.go --cleanup --folder your-folder --vector-store-id <VECTOR_STORE_ID> --output manifest_updated.json
```

#### Command-Line Flags
`--folder`: Folder to scan for files (default: ./your-folder).
`--vector-store-id`: ID of the OpenAI Vector Store.
`--cleanup`: Enable cleanup of deleted files in OpenAI.
`--dry-run`: Disable uploading to OpenAI.
`--output`: Output file for the manifest; if not specified, print to console.
