package cmd

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ryogaendo/agora-agent/internal/config"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/spf13/cobra"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage skills on Anthropic Managed Agents",
}

var skillSyncCmd = &cobra.Command{
	Use:   "sync [skill-name...]",
	Short: "Upload organon skills to Anthropic",
	Long: `Upload skills from organon/skills/ to Anthropic Managed Agents.
If no names are specified, all skills are synced.

Examples:
  agora-agent skill sync                    # sync all skills
  agora-agent skill sync theoria techne     # sync specific skills
  agora-agent skill list                    # list uploaded skills`,
	RunE: runSkillSync,
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List uploaded skills",
	RunE:  runSkillList,
}

func init() {
	skillCmd.AddCommand(skillSyncCmd)
	skillCmd.AddCommand(skillListCmd)
	rootCmd.AddCommand(skillCmd)
}

func runSkillSync(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := anthropic.NewClient()

	// Find organon skills directory
	home, _ := os.UserHomeDir()
	skillsDir := filepath.Join(home, "projects", "organon", "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		// Fallback to ~/.claude/skills
		skillsDir = filepath.Join(home, ".claude", "skills")
	}

	// Determine which skills to sync
	var skillNames []string
	if len(args) > 0 {
		skillNames = args
	} else {
		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			return fmt.Errorf("reading skills dir %s: %w", skillsDir, err)
		}
		for _, e := range entries {
			if e.IsDir() {
				// Check SKILL.md exists
				if _, err := os.Stat(filepath.Join(skillsDir, e.Name(), "SKILL.md")); err == nil {
					skillNames = append(skillNames, e.Name())
				}
			}
		}
	}

	if len(skillNames) == 0 {
		fmt.Println("No skills found to sync.")
		return nil
	}

	fmt.Printf("Syncing %d skill(s) from %s\n\n", len(skillNames), skillsDir)

	uploadedSkills := make(map[string]string) // name -> skill_id
	for _, name := range skillNames {
		skillDir := filepath.Join(skillsDir, name)
		if _, err := os.Stat(skillDir); os.IsNotExist(err) {
			fmt.Printf("  ✗ %s — directory not found\n", name)
			continue
		}

		// Collect all files in the skill directory
		var files []fileEntry
		err := filepath.WalkDir(skillDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
				return nil
			}
			relPath, _ := filepath.Rel(skillDir, path)
			files = append(files, fileEntry{path: path, relPath: relPath})
			return nil
		})
		if err != nil {
			fmt.Printf("  ✗ %s — error walking dir: %v\n", name, err)
			continue
		}

		// Open file readers
		var readers []fileReader
		for _, f := range files {
			file, err := os.Open(f.path)
			if err != nil {
				fmt.Printf("  ✗ %s — error opening %s: %v\n", name, f.relPath, err)
				continue
			}
			readers = append(readers, fileReader{file: file, name: filepath.Join(name, f.relPath)})
		}

		// Build io.Reader slice with named files
		ioReaders := make([]interface{ Read([]byte) (int, error) }, len(readers))
		for i, r := range readers {
			ioReaders[i] = r.file
		}

		// Upload
		resp, err := client.Beta.Skills.New(ctx, anthropic.BetaSkillNewParams{
			DisplayTitle: anthropic.Opt(name),
			Files:        toIOReaders(readers),
		})

		// Close files
		for _, r := range readers {
			r.file.Close()
		}

		if err != nil {
			fmt.Printf("  ✗ %s — upload failed: %v\n", name, err)
			continue
		}

		uploadedSkills[name] = resp.ID
		fmt.Printf("  ✓ %s → %s (version: %s)\n", name, resp.ID, resp.LatestVersion)
	}

	// Save skill IDs to config
	if len(uploadedSkills) > 0 {
		if cfg.Skills == nil {
			cfg.Skills = make(map[string]string)
		}
		for name, id := range uploadedSkills {
			cfg.Skills[name] = id
		}
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("\n✓ %d skill(s) synced. IDs saved to config.\n", len(uploadedSkills))
	}

	return nil
}

func runSkillList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	client := anthropic.NewClient()

	iter := client.Beta.Skills.ListAutoPaging(ctx, anthropic.BetaSkillListParams{})
	count := 0
	for iter.Next() {
		skill := iter.Current()
		fmt.Printf("  %s  %s  (version: %s)\n", skill.ID, skill.DisplayTitle, skill.LatestVersion)
		count++
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if count == 0 {
		fmt.Println("No skills uploaded yet. Run: agora-agent skill sync")
	}

	return nil
}

type fileEntry struct {
	path    string
	relPath string
}

type fileReader struct {
	file *os.File
	name string
}

func toIOReaders(readers []fileReader) []io.Reader {
	result := make([]io.Reader, len(readers))
	for i, r := range readers {
		result[i] = r.file
	}
	return result
}
