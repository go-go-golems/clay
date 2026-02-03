package builder

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// FilterSettings contains all the filter parameters used by the list command.
// These correspond to the filter methods provided by the Builder.
type FilterSettings struct {
	Type          string   `glazed:"type" help:"Filter by command type"`
	Types         []string `glazed:"types" help:"Filter by multiple types (OR)"`
	Tag           string   `glazed:"tag" help:"Filter by single tag"`
	Tags          []string `glazed:"tags" help:"Filter by any of multiple tags (OR)"`
	AllTags       []string `glazed:"all-tags" help:"Must have all specified tags (AND)"`
	AnyTags       []string `glazed:"any-tags" help:"Must have any of specified tags (OR) (alias for --tags)"`
	Path          string   `glazed:"path" help:"Exact path match (e.g., 'queries es')"`
	PathGlob      string   `glazed:"path-glob" help:"Path glob pattern (e.g., 'queries/*')"`
	PathPrefix    string   `glazed:"path-prefix" help:"Path prefix match (e.g., 'queries/')"`
	Name          string   `glazed:"name" help:"Exact command name match (last part of path)"`
	NamePattern   string   `glazed:"name-pattern" help:"Command name pattern match (e.g., 'list*')"`
	MetadataKey   string   `glazed:"metadata-key" help:"Metadata key to match"`
	MetadataValue string   `glazed:"metadata-value" help:"Metadata value to match (requires --metadata-key)"`
}

// FilterSectionSlug is the slug for the filter section.
const FilterSectionSlug = "filter"

// NewFilterSection creates a new section for command filtering.
func NewFilterSection(options ...schema.SectionOption) (schema.Section, error) {
	return schema.NewSection(FilterSectionSlug, "Command Filtering Options",
		append([]schema.SectionOption{
			schema.WithFields(
				fields.New(
					"type",
					fields.TypeString,
					fields.WithHelp("Filter by command type"),
				),
				fields.New(
					"types",
					fields.TypeStringList,
					fields.WithHelp("Filter by multiple types (OR)"),
				),
				fields.New(
					"tag",
					fields.TypeString,
					fields.WithHelp("Filter by single tag"),
				),
				fields.New(
					"tags",
					fields.TypeStringList,
					fields.WithHelp("Filter by any of multiple tags (OR)"),
				),
				fields.New(
					"all-tags",
					fields.TypeStringList,
					fields.WithHelp("Must have all specified tags (AND)"),
				),
				fields.New(
					"any-tags",
					fields.TypeStringList,
					fields.WithHelp("Must have any of specified tags (OR) (alias for --tags)"),
				),
				fields.New(
					"path",
					fields.TypeString,
					fields.WithHelp("Exact path match (e.g., 'queries es')"),
				),
				fields.New(
					"path-glob",
					fields.TypeString,
					fields.WithHelp("Path glob pattern (e.g., 'queries/*')"),
				),
				fields.New(
					"path-prefix",
					fields.TypeString,
					fields.WithHelp("Path prefix match (e.g., 'queries/')"),
				),
				fields.New(
					"name",
					fields.TypeString,
					fields.WithHelp("Exact command name match (last part of path)"),
				),
				fields.New(
					"name-pattern",
					fields.TypeString,
					fields.WithHelp("Command name pattern match (e.g., 'list*')"),
				),
				fields.New(
					"metadata-key",
					fields.TypeString,
					fields.WithHelp("Metadata key to match"),
				),
				fields.New(
					"metadata-value",
					fields.TypeString,
					fields.WithHelp("Metadata value to match (requires --metadata-key)"),
				),
			),
		}, options...)...,
	)
}

// GetFilterSettingsFromParsedValues extracts filter settings from parsed values.
func GetFilterSettingsFromParsedValues(parsedValues *values.Values) (*FilterSettings, error) {
	s := &FilterSettings{}
	err := parsedValues.DecodeSectionInto(FilterSectionSlug, s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize filter settings")
	}
	return s, nil
}

// BuildFilterFromSettings creates a filter from the settings provided via command-line flags.
// It uses the Builder to construct the appropriate query.
func BuildFilterFromSettings(s *FilterSettings, b *Builder) *FilterBuilder {
	var filters []*FilterBuilder

	// Add type filters
	if s.Type != "" {
		filters = append(filters, b.Type(s.Type))
	}
	if len(s.Types) > 0 {
		filters = append(filters, b.Types(s.Types...))
	}

	// Add tag filters
	if s.Tag != "" {
		filters = append(filters, b.Tag(s.Tag))
	}
	// Combine Tags and AnyTags as they are aliases
	tags := s.Tags
	if len(s.AnyTags) > 0 {
		tags = append(tags, s.AnyTags...)
	}
	if len(tags) > 0 {
		// Remove duplicates if any
		uniqueTags := map[string]struct{}{}
		for _, tag := range tags {
			uniqueTags[tag] = struct{}{}
		}
		finalTags := make([]string, 0, len(uniqueTags))
		for tag := range uniqueTags {
			finalTags = append(finalTags, tag)
		}
		if len(finalTags) > 0 {
			filters = append(filters, b.Tags(finalTags...))
		}
	}

	if len(s.AllTags) > 0 {
		filters = append(filters, b.AllTags(s.AllTags...))
	}

	// Add path filters
	if s.Path != "" {
		filters = append(filters, b.Path(s.Path))
	}
	if s.PathGlob != "" {
		filters = append(filters, b.PathGlob(s.PathGlob))
	}
	if s.PathPrefix != "" {
		filters = append(filters, b.PathPrefix(s.PathPrefix))
	}

	// Add name filters
	if s.Name != "" {
		filters = append(filters, b.Name(s.Name))
	}
	if s.NamePattern != "" {
		filters = append(filters, b.NamePattern(s.NamePattern))
	}

	// Add metadata filters
	if s.MetadataKey != "" && s.MetadataValue != "" {
		// TODO(manuel, 2024-07-18) Consider supporting non-string metadata values later
		filters = append(filters, b.Metadata(s.MetadataKey, s.MetadataValue))
	} else if s.MetadataKey != "" || s.MetadataValue != "" {
		// Warn if only one metadata flag is provided?
		log.Warn().Msg("Both --metadata-key and --metadata-value must be provided for metadata filtering")
	}

	// Combine all filters with AND logic
	if len(filters) == 0 {
		// No filters specified, match everything
		return b.MatchAll()
	}

	// Start with the first filter
	result := filters[0]

	// AND it with the rest
	for i := 1; i < len(filters); i++ {
		result = result.And(filters[i])
	}

	return result
}
