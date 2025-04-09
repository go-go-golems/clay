package profiles

import (
	"fmt"

	yaml_editor "github.com/go-go-golems/clay/pkg/yaml-editor"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type ProfileName = string
type LayerName = string
type SettingName = string
type SettingValue = string

type LayerSettings = *orderedmap.OrderedMap[SettingName, SettingValue]
type ProfileLayers = *orderedmap.OrderedMap[LayerName, LayerSettings]
type Profiles = *orderedmap.OrderedMap[ProfileName, ProfileLayers]

type ProfilesEditor struct {
	editor *yaml_editor.YAMLEditor
	path   string
}

func NewProfilesEditor(path string) (*ProfilesEditor, error) {
	editor, err := yaml_editor.NewYAMLEditorFromFile(path)
	if err != nil {
		// Handle case where file doesn't exist? The yaml-editor likely does this.
		// If the file doesn't exist, NewYAMLEditorFromFile might return an empty editor,
		// which should be fine for operations like SetNode (which creates paths).
		// Let's assume the underlying editor handles this for now.
		// TODO(manuel, 2024-07-17) Verify how yaml-editor handles non-existent files
		return nil, fmt.Errorf("could not create editor for %s: %w", path, err)
	}

	return &ProfilesEditor{
		editor: editor,
		path:   path,
	}, nil
}

func (p *ProfilesEditor) Save() error {
	return p.editor.Save(p.path)
}

func (p *ProfilesEditor) SetLayerValue(profile, layer, key, value string) error {
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
	}
	// This should create intermediate nodes if they don't exist.
	return p.editor.SetNode(valueNode, profile, layer, key)
}

func (p *ProfilesEditor) GetLayerValue(profile, layer, key string) (string, error) {
	node, err := p.editor.GetNode(profile, layer, key)
	if err != nil {
		return "", fmt.Errorf("could not get value for %s.%s.%s: %w", profile, layer, key, err)
	}
	return node.Value, nil
}

func (p *ProfilesEditor) DeleteProfile(profile string) error {
	// Setting node to nil should effectively delete it.
	return p.editor.SetNode(nil, profile)
}

func (p *ProfilesEditor) DeleteLayer(profile, layer string) error {
	// Setting node to nil should effectively delete it.
	return p.editor.SetNode(nil, profile, layer)
}

func (p *ProfilesEditor) DeleteLayerValue(profile, layer, key string) error {
	// Setting node to nil should effectively delete it.
	return p.editor.SetNode(nil, profile, layer, key)
}

// ListProfiles returns the names of all profiles and their full content.
func (p *ProfilesEditor) ListProfiles() ([]ProfileName, map[ProfileName]map[LayerName]map[SettingName]SettingValue, error) {
	root, err := p.editor.GetNode()
	if err != nil {
		// If the file was empty or didn't exist initially, GetNode might return an error
		// or an empty/nil node. Treat this as no profiles.
		// Need to check how yaml-editor signals this. Assuming an error for now.
		// If it's a specific "not found" error type, we could handle it more gracefully.
		// Or if root is just nil/empty, handle that.
		// Let's assume for now GetNode returns an error or an empty mapping node.
		if root == nil || len(root.Content) == 0 {
			return []ProfileName{}, make(map[ProfileName]map[LayerName]map[SettingName]SettingValue), nil
		}
		// If it's another error, return it
		return nil, nil, fmt.Errorf("could not get root node: %w", err)

	}

	if root.Kind != yaml.MappingNode {
		// If the root exists but isn't a map (e.g., just a scalar or sequence), it's invalid.
		return nil, nil, fmt.Errorf("root node is not a mapping")
	}

	profiles := make([]ProfileName, 0)
	profileContents := make(map[ProfileName]map[LayerName]map[SettingName]SettingValue)

	for i := 0; i < len(root.Content); i += 2 {
		profileNameNode := root.Content[i]
		profileContentNode := root.Content[i+1]

		if profileNameNode.Kind != yaml.ScalarNode {
			// Skip invalid entries where key is not a scalar
			continue
		}
		profileName := profileNameNode.Value

		// Check if the value node is a mapping node before proceeding
		if profileContentNode.Kind != yaml.MappingNode {
			// Store profile name but indicate empty/invalid content
			profiles = append(profiles, profileName)
			profileContents[profileName] = make(map[LayerName]map[SettingName]SettingValue)
			continue
		}

		// Get the full content for this profile
		layers, err := p.decodeProfileLayers(profileContentNode)
		if err != nil {
			// Log or handle error? For now, just return the error.
			return nil, nil, fmt.Errorf("could not decode layers for profile %s: %w", profileName, err)
		}

		profiles = append(profiles, profileName)

		// Convert the ordered maps to regular maps for the return value
		profileContents[profileName] = make(map[LayerName]map[SettingName]SettingValue)
		for pair := layers.Oldest(); pair != nil; pair = pair.Next() {
			layerName := pair.Key
			settings := pair.Value

			profileContents[profileName][layerName] = make(map[SettingName]SettingValue)
			for settingPair := settings.Oldest(); settingPair != nil; settingPair = settingPair.Next() {
				profileContents[profileName][layerName][settingPair.Key] = settingPair.Value
			}
		}
	}

	return profiles, profileContents, nil
}

// GetProfileLayers returns the layers for a specific profile as an ordered map.
func (p *ProfilesEditor) GetProfileLayers(profile ProfileName) (ProfileLayers, error) {
	profileNode, err := p.editor.GetNode(profile)
	if err != nil {
		return nil, fmt.Errorf("could not get profile '%s': %w", profile, err)
	}

	if profileNode.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("profile '%s' node is not a mapping", profile)
	}

	return p.decodeProfileLayers(profileNode)
}

// decodeProfileLayers takes a yaml.Node (expected to be a MappingNode representing a profile)
// and decodes its content into an ordered map of layers and settings.
func (p *ProfilesEditor) decodeProfileLayers(profileNode *yaml.Node) (ProfileLayers, error) {
	if profileNode.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected a mapping node for profile layers")
	}

	layers := orderedmap.New[LayerName, LayerSettings]()
	for i := 0; i < len(profileNode.Content); i += 2 {
		layerNameNode := profileNode.Content[i]
		layerContentNode := profileNode.Content[i+1]

		if layerNameNode.Kind != yaml.ScalarNode {
			// Skip non-scalar layer names
			continue
		}
		layerName := layerNameNode.Value

		if layerContentNode.Kind != yaml.MappingNode {
			// Layer content is not a map, store layer name with empty settings
			layers.Set(layerName, orderedmap.New[SettingName, SettingValue]())
			continue
		}

		settings := orderedmap.New[SettingName, SettingValue]()
		for j := 0; j < len(layerContentNode.Content); j += 2 {
			keyNode := layerContentNode.Content[j]
			valueNode := layerContentNode.Content[j+1]

			if keyNode.Kind != yaml.ScalarNode || valueNode.Kind != yaml.ScalarNode {
				// Skip non-scalar key/value pairs
				continue
			}
			key := keyNode.Value
			value := valueNode.Value
			settings.Set(key, value)
		}
		layers.Set(layerName, settings)
	}

	return layers, nil
}

// DuplicateProfile copies an existing profile to a new name.
func (p *ProfilesEditor) DuplicateProfile(sourceProfile, newProfile string) error {
	// Get the source profile node
	sourceNode, err := p.editor.GetNode(sourceProfile)
	if err != nil {
		return fmt.Errorf("could not get source profile '%s': %w", sourceProfile, err)
	}

	// Check if target profile already exists
	existingNode, err := p.editor.GetNode(newProfile)
	if err == nil && existingNode != nil {
		return fmt.Errorf("profile '%s' already exists", newProfile)
	}
	// If GetNode returned an error, that's expected if the profile doesn't exist.
	// We only care if it *succeeded* and found an existing node (handled above).
	// If there was a different kind of error (e.g., file read error),
	// the subsequent SetNode call will likely fail anyway.

	// Create a deep copy of the source node. yaml-editor might have a helper?
	// Manual deep copy for now. This is brittle if node structure changes.
	newNode := deepCopyNode(sourceNode)
	if newNode == nil {
		return fmt.Errorf("failed to deep copy source profile node") // Should not happen if sourceNode exists
	}

	// Set the new profile using the copied node
	if err := p.editor.SetNode(newNode, newProfile); err != nil {
		return fmt.Errorf("could not set new profile '%s': %w", newProfile, err)
	}

	return nil
}

// deepCopyNode creates a deep copy of a yaml.Node.
// NOTE: This is a basic implementation and might miss edge cases (e.g., complex aliases/anchors).
func deepCopyNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}

	copy_ := &yaml.Node{
		Kind:        node.Kind,
		Style:       node.Style,
		Tag:         node.Tag,
		Value:       node.Value,
		Anchor:      node.Anchor, // Be careful with anchors/aliases on deep copy
		Alias:       deepCopyNode(node.Alias),
		HeadComment: node.HeadComment,
		LineComment: node.LineComment,
		FootComment: node.FootComment,
		Line:        node.Line,   // Line/Column might not make sense on a copy
		Column:      node.Column, // Line/Column might not make sense on a copy
	}

	if node.Content != nil {
		copy_.Content = make([]*yaml.Node, len(node.Content))
		for i, child := range node.Content {
			copy_.Content[i] = deepCopyNode(child)
		}
	}

	return copy_
}
