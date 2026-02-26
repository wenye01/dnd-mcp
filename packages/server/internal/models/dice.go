package models

// CritStatus 暴击状态
// 规则参考: PHB 第9章 Combat / Critical Hits
type CritStatus string

const (
	// CritStatusNone 普通结果
	CritStatusNone CritStatus = "none"
	// CritStatusSuccess 暴击（自然20）
	CritStatusSuccess CritStatus = "critical"
	// CritStatusFail 大失败（自然1）
	CritStatusFail CritStatus = "fumble"
)

// DiceResult 骰子投掷结果
type DiceResult struct {
	Formula    string     `json:"formula"`     // 骰子公式（如 "1d20+5"）
	Rolls      []int      `json:"rolls"`       // 原始投掷值
	Modifier   int        `json:"modifier"`    // 修正值
	Total      int        `json:"total"`       // 总计
	CritStatus CritStatus `json:"crit_status"` // 暴击状态
}

// NewDiceResult 创建骰子结果
func NewDiceResult(formula string) *DiceResult {
	return &DiceResult{
		Formula:    formula,
		Rolls:      make([]int, 0),
		Modifier:   0,
		Total:      0,
		CritStatus: CritStatusNone,
	}
}

// AddRoll 添加投掷值
func (d *DiceResult) AddRoll(roll int) {
	d.Rolls = append(d.Rolls, roll)
	d.recalculateTotal()
}

// SetRolls 设置所有投掷值
func (d *DiceResult) SetRolls(rolls []int) {
	d.Rolls = rolls
	d.recalculateTotal()
}

// SetModifier 设置修正值
func (d *DiceResult) SetModifier(modifier int) {
	d.Modifier = modifier
	d.recalculateTotal()
}

// recalculateTotal 重新计算总计
func (d *DiceResult) recalculateTotal() {
	sum := 0
	for _, roll := range d.Rolls {
		sum += roll
	}
	d.Total = sum + d.Modifier
}

// CheckCritStatus 检查暴击状态
// 仅对 d20 检定有效：自然1为大失败，自然20为暴击
// 规则参考: PHB 第9章 Combat / Critical Hits
func (d *DiceResult) CheckCritStatus() {
	// 只检查第一个骰子的自然值（通常是 d20）
	if len(d.Rolls) == 1 {
		roll := d.Rolls[0]
		if roll == 20 {
			d.CritStatus = CritStatusSuccess
		} else if roll == 1 {
			d.CritStatus = CritStatusFail
		}
	}
}

// IsCritical 是否暴击
func (d *DiceResult) IsCritical() bool {
	return d.CritStatus == CritStatusSuccess
}

// IsFumble 是否大失败
func (d *DiceResult) IsFumble() bool {
	return d.CritStatus == CritStatusFail
}

// CheckResult 检定结果
type CheckResult struct {
	DiceResult *DiceResult `json:"dice_result"` // 骰子结果
	Ability    string      `json:"ability"`     // 检定属性
	Skill      string      `json:"skill"`       // 技能（可选）
	DC         int         `json:"dc"`          // 难度等级
	Success    bool        `json:"success"`     // 是否成功
	Margin     int         `json:"margin"`      // 成功/失败幅度
}

// NewCheckResult 创建检定结果
func NewCheckResult(diceResult *DiceResult, ability string) *CheckResult {
	return &CheckResult{
		DiceResult: diceResult,
		Ability:    ability,
		DC:         0,
		Success:    false,
		Margin:     0,
	}
}

// SetSkill 设置技能
func (c *CheckResult) SetSkill(skill string) {
	c.Skill = skill
}

// SetDC 设置难度等级并计算成功与否
func (c *CheckResult) SetDC(dc int) {
	c.DC = dc
	c.Evaluate()
}

// Evaluate 评估检定结果
// 规则参考: PHB 第7章 Ability Checks
func (c *CheckResult) Evaluate() {
	if c.DiceResult == nil {
		return
	}

	// 计算幅度
	c.Margin = c.DiceResult.Total - c.DC

	// 自然20总是成功，自然1总是失败（可选规则，某些DM可能不使用）
	// 根据标准规则，只检查是否达到DC
	c.Success = c.DiceResult.Total >= c.DC
}

// IsSuccess 是否成功
func (c *CheckResult) IsSuccess() bool {
	return c.Success
}

// IsFailure 是否失败
func (c *CheckResult) IsFailure() bool {
	return !c.Success
}

// GetMargin 获取成功/失败幅度
func (c *CheckResult) GetMargin() int {
	return c.Margin
}

// RollType 投掷类型
type RollType string

const (
	// RollTypeNormal 正常投掷
	RollTypeNormal RollType = "normal"
	// RollTypeAdvantage 优势
	RollTypeAdvantage RollType = "advantage"
	// RollTypeDisadvantage 劣势
	RollTypeDisadvantage RollType = "disadvantage"
)

// DiceFormula 骰子公式解析结果
type DiceFormula struct {
	Count    int      // 骰子数量
	Sides    int      // 骰子面数
	Modifier int      // 修正值
	KeepHigh int      // 保留最高N个（0表示保留全部）
	KeepLow  int      // 保留最低N个（0表示保留全部）
	RollType RollType // 投掷类型
	Original string   // 原始公式字符串
}

// NewDiceFormula 创建默认骰子公式
func NewDiceFormula() *DiceFormula {
	return &DiceFormula{
		Count:    1,
		Sides:    20,
		Modifier: 0,
		KeepHigh: 0,
		KeepLow:  0,
		RollType: RollTypeNormal,
	}
}

// String 返回公式字符串表示
func (f *DiceFormula) String() string {
	result := ""
	if f.RollType == RollTypeAdvantage || f.RollType == RollTypeDisadvantage {
		// 优势/劣势的特殊表示
		if f.RollType == RollTypeAdvantage {
			result = "max(2d20)"
		} else {
			result = "min(2d20)"
		}
		if f.Modifier != 0 {
			if f.Modifier > 0 {
				result += "+"
			}
			result += string(rune(f.Modifier + '0'))
		}
	} else {
		result = string(rune(f.Count+'0')) + "d" + string(rune(f.Sides+'0'))
		if f.KeepHigh > 0 {
			result += "kh" + string(rune(f.KeepHigh+'0'))
		}
		if f.KeepLow > 0 {
			result += "kl" + string(rune(f.KeepLow+'0'))
		}
		if f.Modifier != 0 {
			if f.Modifier > 0 {
				result += "+"
			}
			result += string(rune(f.Modifier + '0'))
		}
	}
	return result
}

// IsKeepRoll 是否为保留骰（kh/kl）
func (f *DiceFormula) IsKeepRoll() bool {
	return f.KeepHigh > 0 || f.KeepLow > 0
}

// IsAdvantage 是否为优势
func (f *DiceFormula) IsAdvantage() bool {
	return f.RollType == RollTypeAdvantage
}

// IsDisadvantage 是否为劣势
func (f *DiceFormula) IsDisadvantage() bool {
	return f.RollType == RollTypeDisadvantage
}
