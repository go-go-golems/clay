# YAML Editor Integration in Repository Commands

Updated the repository management commands to use the new YAML editor:
- Refactored add, remove, and get commands to use the YAML editor
- Improved error handling and reporting
- Removed manual YAML manipulation code
- Preserved comments and formatting in config files
- Simplified code by leveraging editor's helper functions 

# Command Filter Integration in Sqleton

Added the filter command to sqleton under the new "commands" subgroup:
- Created new "commands" subgroup for command management
- Added filter command for searching and filtering sqleton commands
- Added helper function to convert commands to descriptions

# Command Filter Implementation

Added a new filter command that allows filtering commands based on various criteria like type, tags, path, etc. The command uses Bleve for efficient searching and supports complex filtering with pattern matching and metadata search.

- Added `filter` command in `pkg/filters/command/cmds`
- Added `MatchAll` method to command filter builder
- Supports filtering by type, tags, path, name, and metadata
- Outputs results in a structured format using glazed 