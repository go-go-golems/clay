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
9. [ ] Walk an empty directory
10. [ ] Walk a directory with only files (no subdirectories)
11. [ ] Walk a directory with only subdirectories (no files)
12. [ ] Handle deeply nested directory structures
13. [ ] Handle files and directories with special characters in names
14. [ ] Handle files with very large content
15. [ ] Handle directories with a large number of entries
16. [ ] Walk a directory structure with circular symlinks
17. [ ] Handle Unicode characters in file and directory names
18. [ ] Test with root directory as the starting point
19. [ ] Test with hidden files and directories
20. [ ] Handle files with no extension
21. [ ] Test with read-only filesystems

## Error Cases
22. [ ] Attempt to walk a non-existent path
23. [ ] Attempt to walk a file instead of a directory
24. [ ] Handle permission denied errors when reading directories or files
25. [ ] Handle I/O errors during file reading
26. [ ] Attempt to retrieve a non-existent node by path
27. [ ] Handle out-of-memory scenarios for very large directory structures
28. [ ] Handle filesystem changes during traversal (e.g., file deletion)
29. [ ] Handle errors in pre-visit or post-visit functions
30. [ ] Attempt to walk a path that exceeds maximum path length for the OS

## Performance Tests
31. [ ] Benchmark walking large directory structures
32. [ ] Benchmark retrieving nodes by path in large structures

## Concurrency Tests
33. [ ] Test concurrent access to Walker methods
34. [ ] Test parallel walking of multiple root paths

Note: As tests are implemented, update this list by marking completed items with [x].