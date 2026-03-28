package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PromptType 提示词类型
type PromptType string

const (
	PromptTypeIndicator PromptType = "indicator" // 指标分析
	PromptTypeScreener  PromptType = "screener"  // 选股筛选
	PromptTypeReview    PromptType = "review"    // 复盘分析
	PromptTypePersona   PromptType = "persona"   // AI人设
)

// PromptInfo 提示词信息
type PromptInfo struct {
	Name      string     `json:"name"`      // 名称（文件名，不含扩展名）
	Type      PromptType `json:"type"`      // 类型
	Content   string     `json:"content"`   // 内容
	FilePath  string     `json:"filePath"`  // 文件路径
	CreatedAt time.Time  `json:"createdAt"` // 创建时间
	UpdatedAt time.Time  `json:"updatedAt"` // 更新时间
}

// Manager 提示词管理器
type Manager struct {
	baseDir string
}

// NewManager 创建提示词管理器
func NewManager(baseDir string) (*Manager, error) {
	m := &Manager{
		baseDir: baseDir,
	}

	// 确保目录结构存在
	if err := m.ensureDirectories(); err != nil {
		return nil, err
	}

	return m, nil
}

// ensureDirectories 确保所有类型的目录都存在
func (m *Manager) ensureDirectories() error {
	types := []PromptType{
		PromptTypeIndicator,
		PromptTypeScreener,
		PromptTypeReview,
		PromptTypePersona,
	}

	for _, t := range types {
		dir := filepath.Join(m.baseDir, string(t))
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}
	}

	return nil
}

// GetBaseDir 获取基础目录
func (m *Manager) GetBaseDir() string {
	return m.baseDir
}

// GetTypeDir 获取指定类型的目录
func (m *Manager) GetTypeDir(promptType PromptType) (string, error) {
	if !isValidPromptType(promptType) {
		return "", fmt.Errorf("提示词类型 %q 不受支持", promptType)
	}
	return filepath.Join(m.baseDir, string(promptType)), nil
}

// List 列出指定类型的所有提示词
func (m *Manager) List(promptType PromptType) ([]PromptInfo, error) {
	dir, err := m.GetTypeDir(promptType)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []PromptInfo{}, nil
		}
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	var prompts []PromptInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".txt") {
			continue
		}

		filePath := filepath.Join(dir, name)
		info, err := m.readPromptFile(filePath, promptType)
		if err != nil {
			continue
		}

		prompts = append(prompts, *info)
	}

	return prompts, nil
}

// ListAll 列出所有类型的提示词
func (m *Manager) ListAll() (map[PromptType][]PromptInfo, error) {
	result := make(map[PromptType][]PromptInfo)

	types := []PromptType{
		PromptTypeIndicator,
		PromptTypeScreener,
		PromptTypeReview,
		PromptTypePersona,
	}

	for _, t := range types {
		prompts, err := m.List(t)
		if err != nil {
			return nil, err
		}
		result[t] = prompts
	}

	return result, nil
}

// Get 获取指定提示词
func (m *Manager) Get(promptType PromptType, name string) (*PromptInfo, error) {
	filePath, err := m.getFilePath(promptType, name)
	if err != nil {
		return nil, err
	}
	return m.readPromptFile(filePath, promptType)
}

// Create 创建提示词
func (m *Manager) Create(promptType PromptType, name string, content string) (*PromptInfo, error) {
	// 检查名称是否合法
	if name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}

	// 移除.txt后缀（如果有）
	name = strings.TrimSuffix(name, ".txt")

	filePath, err := m.getFilePath(promptType, name)
	if err != nil {
		return nil, err
	}

	// 检查是否已存在
	if _, err := os.Stat(filePath); err == nil {
		return nil, fmt.Errorf("提示词已存在: %s", name)
	}

	// 写入文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	return m.readPromptFile(filePath, promptType)
}

// Update 更新提示词
func (m *Manager) Update(promptType PromptType, name string, content string) (*PromptInfo, error) {
	name = strings.TrimSuffix(name, ".txt")
	filePath, err := m.getFilePath(promptType, name)
	if err != nil {
		return nil, err
	}

	// 检查是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("提示词不存在: %s", name)
	}

	// 写入文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	return m.readPromptFile(filePath, promptType)
}

// Delete 删除提示词
func (m *Manager) Delete(promptType PromptType, name string) error {
	name = strings.TrimSuffix(name, ".txt")
	filePath, err := m.getFilePath(promptType, name)
	if err != nil {
		return err
	}

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("提示词不存在: %s", name)
		}
		return fmt.Errorf("删除失败: %w", err)
	}

	return nil
}

// Rename 重命名提示词
func (m *Manager) Rename(promptType PromptType, oldName string, newName string) error {
	oldName = strings.TrimSuffix(oldName, ".txt")
	newName = strings.TrimSuffix(newName, ".txt")

	if newName == "" {
		return fmt.Errorf("新名称不能为空")
	}

	oldPath, err := m.getFilePath(promptType, oldName)
	if err != nil {
		return err
	}
	newPath, err := m.getFilePath(promptType, newName)
	if err != nil {
		return err
	}

	// 检查旧文件是否存在
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("提示词不存在: %s", oldName)
	}

	// 检查新文件是否已存在
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("目标名称已存在: %s", newName)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("重命名失败: %w", err)
	}

	return nil
}

// Import 导入提示词（从内容创建）
func (m *Manager) Import(promptType PromptType, name string, content string) (*PromptInfo, error) {
	return m.Create(promptType, name, content)
}

// Export 导出提示词（返回内容）
func (m *Manager) Export(promptType PromptType, name string) (string, error) {
	info, err := m.Get(promptType, name)
	if err != nil {
		return "", err
	}
	return info.Content, nil
}

// getFilePath 获取文件路径
func (m *Manager) getFilePath(promptType PromptType, name string) (string, error) {
	dir, err := m.GetTypeDir(promptType)
	if err != nil {
		return "", err
	}
	cleanName := strings.TrimSuffix(name, ".txt")
	safeName, err := sanitizePromptName(cleanName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, safeName+".txt"), nil
}

func sanitizePromptName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("提示词名称不能为空")
	}
	if filepath.IsAbs(trimmed) {
		return "", fmt.Errorf("提示词名称不能包含路径")
	}
	if strings.Contains(trimmed, "..") {
		return "", fmt.Errorf("提示词名称不能包含路径")
	}
	if strings.ContainsAny(trimmed, `/\\`) {
		return "", fmt.Errorf("提示词名称不能包含路径分隔符")
	}
	if strings.ContainsAny(trimmed, ":*?\"<>|") {
		return "", fmt.Errorf("提示词名称包含非法字符")
	}
	return trimmed, nil
}

// readPromptFile 读取提示词文件
func (m *Manager) readPromptFile(filePath string, promptType PromptType) (*PromptInfo, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	name := strings.TrimSuffix(filepath.Base(filePath), ".txt")

	return &PromptInfo{
		Name:      name,
		Type:      promptType,
		Content:   string(content),
		FilePath:  filePath,
		CreatedAt: fileInfo.ModTime(), // 使用修改时间作为创建时间（无法获取真实创建时间）
		UpdatedAt: fileInfo.ModTime(),
	}, nil
}

func isValidPromptType(promptType PromptType) bool {
	switch promptType {
	case PromptTypeIndicator, PromptTypeScreener, PromptTypeReview, PromptTypePersona:
		return true
	default:
		return false
	}
}

// GetPromptTypes 获取所有提示词类型信息
func GetPromptTypes() []struct {
	Type        PromptType `json:"type"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
} {
	return []struct {
		Type        PromptType `json:"type"`
		Name        string     `json:"name"`
		Description string     `json:"description"`
	}{
		{PromptTypeIndicator, "指标分析", "分析股票技术指标，如MACD、KDJ等，给出买卖信号"},
		{PromptTypeScreener, "选股筛选", "根据条件从股票池中筛选符合要求的股票"},
		{PromptTypeReview, "复盘分析", "分析持仓表现，生成每日/每周复盘报告"},
		{PromptTypePersona, "AI人设", "自定义AI助手的回答风格和专业领域"},
	}
}
