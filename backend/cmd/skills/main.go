package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/skillsmanager/skillsmanager/backend/pkg/api"
	"github.com/skillsmanager/skillsmanager/backend/pkg/models"
)

var configPath string

func main() {
	root := &cobra.Command{
		Use:   "skills",
		Short: "Agent Skills Manager - 跨 Agent 技能安装与管理",
		Long:  `统一管理多 Agent（Trae / Claude / Cursor 等）的技能安装、版本管理与分发。`,
	}

	root.PersistentFlags().StringVar(&configPath, "config", "", "配置文件路径（默认使用系统路径）")

	root.AddCommand(cmdFind())
	root.AddCommand(cmdInstall())
	root.AddCommand(cmdList())
	root.AddCommand(cmdAgents())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cmdFind() *cobra.Command {
	return &cobra.Command{
		Use:   "find [keyword]",
		Short: "搜索技能（提供 GitHub URL 或本地路径即可安装）",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			keyword := ""
			if len(args) > 0 {
				keyword = args[0]
			}
			fmt.Println("ℹ  目前 skills find 支持以下安装方式:")
			fmt.Println()
			fmt.Println("  1) GitHub 仓库 URL:")
			fmt.Println("     skills install https://github.com/user/repo")
			fmt.Println()
			fmt.Println("  2) GitHub 子目录:")
			fmt.Println("     skills install https://github.com/user/repo --sub-path skills/my-skill")
			fmt.Println()
			fmt.Println("  3) 本地目录:")
			fmt.Println("     skills install /path/to/local/skill")
			fmt.Println()
			fmt.Println("  4) 指定分支/标签:")
			fmt.Println("     skills install https://github.com/user/repo --ref main")
			fmt.Println()
			if keyword != "" {
				fmt.Printf("🔍 提示: '%s' - 请确认对应的 GitHub 仓库 URL 或本地路径\n", keyword)
			}
		},
	}
}

func cmdInstall() *cobra.Command {
	var subPath, ver, ref, scope, projectDir string
	var agents []string
	var jsonOut bool

	c := &cobra.Command{
		Use:   "install <source>",
		Short: "安装技能到一个或多个 Agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := api.New(configPath)
			if err != nil {
				return err
			}
			req := api.InstallRequest{
				Source:     args[0],
				SubPath:    subPath,
				Version:    ver,
				Ref:        ref,
				Agents:     agents,
				Scope:      scope,
				ProjectDir: projectDir,
			}
			results, err := a.Install(req)
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(results)
			}
			for _, r := range results {
				fmt.Printf("✅ 已安装: %s @ %s\n", r.SkillName, r.Version)
				fmt.Printf("   来源: %s\n", formatSource(r.Source))
				ids := make([]string, 0, len(r.Agents))
				for id := range r.Agents {
					ids = append(ids, id)
				}
				sort.Strings(ids)
				for _, agentID := range ids {
					link := r.Agents[agentID]
					status := "✅"
					if !link.Success {
						status = "❌ " + link.Error
					}
					fmt.Printf("   %s %s → %s\n", status, agentID, link.Path)
				}
			}
			fmt.Printf("\n📁 skillspool: %s\n", a.SkillspoolRoot())
			return nil
		},
	}
	c.Flags().StringVar(&subPath, "sub-path", "", "仓库子目录")
	c.Flags().StringVar(&ver, "version", "", "强制指定版本号")
	c.Flags().StringVar(&ref, "ref", "", "git 分支/标签")
	c.Flags().StringSliceVar(&agents, "agent", nil, "目标 Agent ID（可重复指定，默认全部已检测的）")
	c.Flags().StringVar(&scope, "scope", "global", "安装范围: global / project")
	c.Flags().StringVar(&projectDir, "project-dir", "", "项目目录（scope=project 时）")
	c.Flags().BoolVar(&jsonOut, "json", false, "JSON 输出")
	return c
}

func cmdList() *cobra.Command {
	var jsonOut bool
	c := &cobra.Command{
		Use:   "list",
		Short: "列出已安装技能",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := api.New(configPath)
			if err != nil {
				return err
			}
			skills := a.ListSkills()
			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(skills)
			}
			if len(skills) == 0 {
				fmt.Println("（暂无已安装技能）")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "技能名\t最新版本\t来源\t分发到")
			for _, s := range skills {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					s.Name, s.LatestVersion, shortSource(s.Source), collectAgents(s))
			}
			w.Flush()
			fmt.Printf("\n📁 skillspool: %s\n", a.SkillspoolRoot())
			return nil
		},
	}
	c.Flags().BoolVar(&jsonOut, "json", false, "JSON 输出")
	return c
}

func cmdAgents() *cobra.Command {
	var jsonOut bool
	c := &cobra.Command{
		Use:   "agents",
		Short: "列出已配置与已检测的 Agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := api.New(configPath)
			if err != nil {
				return err
			}
			agents := a.ListAgents()
			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(agents)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\t名称\t已检测\t全局目录")
			ids := make([]string, 0, len(agents))
			for id := range agents {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			for _, id := range ids {
				ag := agents[id]
				detected := "✅"
				if !ag.Detected {
					detected = "—"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, ag.Name, detected, ag.GlobalLocation)
			}
			w.Flush()
			return nil
		},
	}
	c.Flags().BoolVar(&jsonOut, "json", false, "JSON 输出")
	return c
}

// --- 辅助 ---

func formatSource(s models.Source) string {
	if s.URL != "" {
		return s.URL
	}
	if s.Path != "" {
		return "local: " + s.Path
	}
	if s.Command != "" {
		return s.Command
	}
	return "-"
}

func shortSource(s models.Source) string {
	f := formatSource(s)
	if len(f) > 50 {
		return f[:47] + "..."
	}
	return f
}

func collectAgents(s *models.Skill) string {
	m := map[string]bool{}
	for _, sv := range s.Versions {
		for _, ag := range sv.Agents {
			m[ag] = true
		}
	}
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		return "-"
	}
	return join(", ", ids)
}

func join(sep string, items []string) string {
	if len(items) == 0 {
		return ""
	}
	out := items[0]
	for _, it := range items[1:] {
		out += sep + it
	}
	return out
}
