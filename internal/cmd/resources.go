package cmd

import (
	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/output"
)

// resourceDefinitions lists all v2 API resource commands.
var resourceDefinitions = []ResourceDef{
	{
		Name: "project", Plural: "projects", APIPath: "/v2/projects/", IDField: "id",
		ListCols:  []output.Column{{Header: "ID", Field: "id"}, {Header: "Title", Field: "title"}, {Header: "Category", Field: "category"}, {Header: "Created", Field: "created"}},
		HasCreate: true, HasUpdate: true, HasDelete: true,
	},
	{
		Name: "task", Plural: "tasks", APIPath: "/v2/tasks/", IDField: "id",
		ListCols:  []output.Column{{Header: "ID", Field: "id"}, {Header: "Name", Field: "name"}, {Header: "Status", Field: "status"}},
		HasCreate: true, HasUpdate: true, HasDelete: true,
	},
	{
		Name: "assignment", Plural: "assignments", APIPath: "/v2/assignments/", IDField: "id",
		ListCols: []output.Column{{Header: "ID", Field: "id"}, {Header: "User", Field: "user_id"}, {Header: "Task", Field: "task_id"}, {Header: "Status", Field: "status_label"}},
	},
	{
		Name: "review", Plural: "reviews", APIPath: "/v2/reviews/", IDField: "id",
		ListCols: []output.Column{{Header: "ID", Field: "id"}, {Header: "Status", Field: "status"}},
	},
	{
		Name: "data-collection", Plural: "data collections", APIPath: "/v2/data-collections/", IDField: "id",
		ListCols:  []output.Column{{Header: "ID", Field: "id"}, {Header: "Name", Field: "name"}, {Header: "Created", Field: "created"}},
		HasCreate: true, HasUpdate: true, HasDelete: true,
	},
	{
		Name: "data-file", Plural: "data files", APIPath: "/v2/data-files/", IDField: "id",
		ListCols: []output.Column{{Header: "ID", Field: "id"}, {Header: "Name", Field: "name"}},
	},
	{
		Name: "data-unit", Plural: "data units", APIPath: "/v2/data-units/", IDField: "id",
		ListCols:  []output.Column{{Header: "ID", Field: "id"}, {Header: "Name", Field: "name"}},
		HasCreate: true, HasUpdate: true, HasDelete: true,
	},
	{
		Name: "experiment", Plural: "experiments", APIPath: "/v2/experiments/", IDField: "id",
		ListCols:  []output.Column{{Header: "ID", Field: "id"}, {Header: "Name", Field: "name"}, {Header: "Created", Field: "created"}},
		HasCreate: true, HasUpdate: true, HasDelete: true,
	},
	{
		Name: "gt-dataset", Plural: "ground truth datasets", APIPath: "/v2/ground-truth-datasets/", IDField: "id",
		ListCols:  []output.Column{{Header: "ID", Field: "id"}, {Header: "Name", Field: "name"}},
		HasCreate: true, HasUpdate: true, HasDelete: true,
	},
	{
		Name: "gt", Plural: "ground truths", APIPath: "/v2/ground-truths/", IDField: "id",
		ListCols: []output.Column{{Header: "ID", Field: "id"}},
	},
	{
		Name: "model", Plural: "models", APIPath: "/v2/models/", IDField: "id",
		ListCols: []output.Column{{Header: "ID", Field: "id"}, {Header: "Name", Field: "name"}},
	},
	{
		Name: "job", Plural: "jobs", APIPath: "/v2/jobs/", IDField: "id",
		ListCols: []output.Column{{Header: "ID", Field: "id"}, {Header: "Status", Field: "status"}},
	},
	{
		Name: "plugin", Plural: "plugins", APIPath: "/v2/plugins/", IDField: "id",
		ListCols: []output.Column{{Header: "ID", Field: "id"}, {Header: "Name", Field: "name"}},
	},
	{
		Name: "group", Plural: "groups", APIPath: "/v2/groups/", IDField: "id",
		ListCols:  []output.Column{{Header: "ID", Field: "id"}, {Header: "Name", Field: "name"}},
		HasCreate: true, HasUpdate: true, HasDelete: true,
	},
	{
		Name: "workshop", Plural: "workshops", APIPath: "/v2/workshops/", IDField: "id",
		ListCols: []output.Column{{Header: "ID", Field: "id"}, {Header: "Name", Field: "name"}},
	},
	{
		Name: "member", Plural: "members", APIPath: "/v2/members/", IDField: "id",
		ListCols: []output.Column{{Header: "ID", Field: "id"}, {Header: "Name", Field: "name"}, {Header: "Role", Field: "role"}},
	},
}

// Sub-resources that have permissions/roles/invite patterns
var subresourceMap = map[string][]string{
	"project":         {"permissions", "roles", "invite"},
	"task":            {"permissions", "roles"},
	"assignment":      {"permissions", "roles"},
	"review":          {"permissions", "roles"},
	"data-collection": {"permissions", "roles", "invite"},
	"data-unit":       {"permissions", "roles"},
	"experiment":      {"permissions", "roles", "invite"},
	"gt-dataset":      {"permissions", "roles"},
	"group":           {"permissions", "roles", "invite"},
	"workshop":        {"permissions", "roles"},
}

// registerResourceCommands adds all resource commands to the root command.
func registerResourceCommands(root *cobra.Command) {
	for _, def := range resourceDefinitions {
		cmd := newResourceCmd(def)

		// Add sub-resource commands
		if subs, ok := subresourceMap[def.Name]; ok {
			for _, sub := range subs {
				switch sub {
				case "permissions":
					addPermissionsCmd(cmd, def.APIPath)
				case "roles":
					addRolesCmd(cmd, def.APIPath)
				case "invite":
					addInviteCmd(cmd, def.APIPath)
				}
			}
		}

		// Special sub-commands
		switch def.Name {
		case "job":
			cmd.AddCommand(newResourceCmd(ResourceDef{
				Name: "log", Plural: "job logs", APIPath: "/v2/job-logs/", IDField: "id",
				ListCols: []output.Column{{Header: "ID", Field: "id"}, {Header: "Job", Field: "job_id"}},
			}))
		case "plugin":
			cmd.AddCommand(newResourceCmd(ResourceDef{
				Name: "release", Plural: "plugin releases", APIPath: "/v2/plugin-releases/", IDField: "id",
				ListCols: []output.Column{{Header: "ID", Field: "id"}, {Header: "Version", Field: "version"}},
			}))
		}

		root.AddCommand(cmd)
	}
}
