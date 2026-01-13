package github

// ListLabels returns all labels in a repo
func (c *Client) ListLabels(repo string) ([]LabelInfo, error) {
	var labels []LabelInfo
	err := c.RunJSON(&labels,
		"label", "list",
		"--repo", repo,
		"--json", "name,color,description",
		"--limit", "100",
	)
	return labels, err
}

// CreateLabel creates a new label
func (c *Client) CreateLabel(repo, name, color, description string) error {
	_, err := c.Run(
		"label", "create", name,
		"--repo", repo,
		"--color", color,
		"--description", description,
		"--force", // Update if exists
	)
	return err
}

// LabelExists checks if a label exists
func (c *Client) LabelExists(repo, name string) (bool, error) {
	labels, err := c.ListLabels(repo)
	if err != nil {
		return false, err
	}

	for _, l := range labels {
		if l.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// DeleteLabel deletes a label
func (c *Client) DeleteLabel(repo, name string) error {
	_, err := c.Run(
		"label", "delete", name,
		"--repo", repo,
		"--yes",
	)
	return err
}

// SyncLabels ensures all required labels exist in the repo
func (c *Client) SyncLabels(repo string, labels []LabelInfo) error {
	for _, label := range labels {
		err := c.CreateLabel(repo, label.Name, label.Color, label.Description)
		if err != nil {
			return err
		}
	}
	return nil
}
