# Active Context

## Current Focus
- Implementing and fixing path-based queries in the command filter system
- Ensuring proper text analysis for path fields in Bleve index

## Recent Changes
- Fixed path-based queries in command filter by configuring proper text analysis
- Added keyword analyzer for full_path field to prevent tokenization
- Enhanced debug logging for query construction and execution
- Verified path prefix and glob pattern matching functionality

## Active Decisions
1. Path Field Indexing
   - Using keyword analyzer for full_path field to preserve path structure
   - Paths are stored as complete strings (e.g., "service/api/http-api")
   - No tokenization to maintain path hierarchy

2. Query Construction
   - PathPrefix queries ensure trailing slash for consistency
   - PathGlob queries use wildcard patterns for flexible matching
   - Conjunction queries combine path and type/tag filters

## Next Steps
1. Consider adding more path-based query patterns:
   - Parent path matching
   - Depth-based filtering
   - Multiple path pattern matching

2. Optimization opportunities:
   - Cache common path queries
   - Optimize wildcard pattern matching
   - Add path validation

3. Documentation:
   - Document path query patterns
   - Add examples for common use cases
   - Update API documentation

## Current Considerations
1. Query Performance
   - Monitor performance of wildcard queries
   - Consider indexing strategies for large command sets
   - Evaluate caching options for frequent queries

2. Path Handling
   - Maintain consistent path format
   - Handle edge cases (empty paths, special characters)
   - Consider platform-specific path separators

3. Testing Coverage
   - Add more complex path pattern tests
   - Test edge cases and error conditions
   - Benchmark query performance 