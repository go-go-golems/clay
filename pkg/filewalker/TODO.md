# TODO: Tests for filewalker package

## Happy Path Tests
1. [x] Walk a simple directory structure with files and subdirectories
2. [x] Retrieve a node by its absolute path
3. [x] Retrieve a node by its relative path
4. [x] Follow symlinks when the option is enabled
5. [x] Don't follow symlinks when the option is disabled
6. [x] Correctly handle file metadata (size, permissions, modification time)
7. [x] Traverse multiple root paths
8. [x] Execute pre-visit and post-visit functions correctly

## Edge Cases
9. [x] Walk an empty directory
10. [x] Walk a directory with only files (no subdirectories)
11. [x] Walk a directory with only subdirectories (no files)
12. [x] Handle deeply nested directory structures
13. [x] Handle files and directories with special characters in names
19. [ ] Test with hidden files and directories

## Error Cases
22. [x] Attempt to walk a non-existent path
23. [x] Attempt to walk a file instead of a directory
24. [x] Handle permission denied errors when reading directories or files
25. [x] Handle I/O errors during file reading
26. [x] Attempt to retrieve a non-existent node by path
29. [x] Handle errors in pre-visit or post-visit functions
