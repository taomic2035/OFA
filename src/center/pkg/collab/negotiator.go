// Package collab - Agent协商器
// Sprint 32: v9.0 智能Agent协作
package collab

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Negotiator Agent协商器
type Negotiator struct {
	proposals    map[string]*Proposal
	votes        map[string][]*Vote
	agreements   map[string]*Agreement
	conflictResolver *ConflictResolver
	mu           sync.RWMutex
	timeout      time.Duration
}

// Proposal 提议
type Proposal struct {
	ID          string                 `json:"id"`
	Proposer    string                 `json:"proposer"`    // Agent ID
	Type        ProposalType           `json:"type"`
	Subject     string                 `json:"subject"`     // 协商主题
	Content     map[string]interface{} `json:"content"`     // 提议内容
	Timestamp   time.Time              `json:"timestamp"`
	Status      ProposalStatus         `json:"status"`
	VotesFor    int                    `json:"votes_for"`
	VotesAgainst int                   `json:"votes_against"`
	VotesNeutral int                   `json:"votes_neutral"`
	ExpiresAt   time.Time              `json:"expires_at"`
}

// ProposalType 提议类型
type ProposalType string

const (
	ProposalTaskAssign ProposalType = "task_assign"  // 任务分配
	ProposalResource   ProposalType = "resource"     // 资源分配
	ProposalPolicy     ProposalType = "policy"       // 策略制定
	ProposalPriority   ProposalType = "priority"     // 优先级调整
	ProposalConflict   ProposalType = "conflict"    // 冲突解决
)

// ProposalStatus 提议状态
type ProposalStatus string

const (
	StatusPending   ProposalStatus = "pending"
	StatusVoting    ProposalStatus = "voting"
	StatusApproved  ProposalStatus = "approved"
	StatusRejected  ProposalStatus = "rejected"
	StatusExpired   ProposalStatus = "expired"
)

// Vote 投票
type Vote struct {
	Voter      string    `json:"voter"`      // Agent ID
	ProposalID string    `json:"proposal_id"`
	Decision   VoteDecision `json:"decision"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
	Weight     float64   `json:"weight"`     // 投票权重
}

// VoteDecision 投票决策
type VoteDecision string

const (
	VoteFor     VoteDecision = "for"
	VoteAgainst VoteDecision = "against"
	VoteNeutral VoteDecision = "neutral"
)

// Agreement 协议
type Agreement struct {
	ID          string                 `json:"id"`
	ProposalID  string                 `json:"proposal_id"`
	Parties     []string               `json:"parties"`     // 参与方
	Terms       map[string]interface{} `json:"terms"`       // 协议条款
	Timestamp   time.Time              `json:"timestamp"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Status      AgreementStatus        `json:"status"`
}

// AgreementStatus 协议状态
type AgreementStatus string

const (
	AgreementActive   AgreementStatus = "active"
	AgreementExpired  AgreementStatus = "expired"
	AgreementRevoked  AgreementStatus = "revoked"
)

// ConflictResolver 冲突解决器
type ConflictResolver struct {
	strategies map[ConflictType]ResolutionStrategy
}

// ConflictType 冲突类型
type ConflictType string

const (
	ConflictResource ConflictType = "resource"  // 资源冲突
	ConflictTask     ConflictType = "task"      // 任务冲突
	ConflictPriority ConflictType = "priority"  // 优先级冲突
	ConflictData     ConflictType = "data"      // 数据冲突
)

// ResolutionStrategy 解决策略
type ResolutionStrategy string

const (
	StrategyFirstCome ResolutionStrategy = "first_come"  // 先来先得
	StrategyPriority  ResolutionStrategy = "priority"   // 优先级优先
	StrategyVote      ResolutionStrategy = "vote"       // 投票决定
	StrategyRandom    ResolutionStrategy = "random"     // 随机分配
	StrategyNegotiate ResolutionStrategy = "negotiate"  // 协商解决
)

// NewNegotiator 创建协商器
func NewNegotiator() *Negotiator {
	return &Negotiator{
		proposals:    make(map[string]*Proposal),
		votes:        make(map[string][]*Vote),
		agreements:   make(map[string]*Agreement),
		conflictResolver: NewConflictResolver(),
		timeout:      30 * time.Second,
	}
}

// NewConflictResolver 创建冲突解决器
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{
		strategies: map[ConflictType]ResolutionStrategy{
			ConflictResource: StrategyPriority,
			ConflictTask:     StrategyNegotiate,
			ConflictPriority: StrategyVote,
			ConflictData:     StrategyFirstCome,
		},
	}
}

// SetTimeout 设置超时
func (n *Negotiator) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.timeout = timeout
}

// CreateProposal 创建提议
func (n *Negotiator) CreateProposal(proposer, subject string, proposalType ProposalType, content map[string]interface{}) *Proposal {
	n.mu.Lock()
	defer n.mu.Unlock()

	proposal := &Proposal{
		ID:        generateProposalID(),
		Proposer:  proposer,
		Type:      proposalType,
		Subject:   subject,
		Content:   content,
		Timestamp: time.Now(),
		Status:    StatusVoting,
		ExpiresAt: time.Now().Add(n.timeout),
	}

	n.proposals[proposal.ID] = proposal
	n.votes[proposal.ID] = make([]*Vote, 0)

	return proposal
}

// Vote 提交投票
func (n *Negotiator) Vote(proposalID, voter string, decision VoteDecision, reason string, weight float64) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	proposal, ok := n.proposals[proposalID]
	if !ok {
		return fmt.Errorf("提议不存在: %s", proposalID)
	}

	// 检查是否过期
	if time.Now().After(proposal.ExpiresAt) {
		proposal.Status = StatusExpired
		return fmt.Errorf("提议已过期")
	}

	// 检查是否已投票
	for _, v := range n.votes[proposalID] {
		if v.Voter == voter {
			return fmt.Errorf("已投票: %s", voter)
		}
	}

	vote := &Vote{
		Voter:      voter,
		ProposalID: proposalID,
		Decision:   decision,
		Reason:     reason,
		Timestamp:  time.Now(),
		Weight:     weight,
	}

	n.votes[proposalID] = append(n.votes[proposalID], vote)

	// 更新投票统计
	switch decision {
	case VoteFor:
		proposal.VotesFor++
	case VoteAgainst:
		proposal.VotesAgainst++
	case VoteNeutral:
		proposal.VotesNeutral++
	}

	return nil
}

// ResolveProposal 解决提议
func (n *Negotiator) ResolveProposal(proposalID string) (*Agreement, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	proposal, ok := n.proposals[proposalID]
	if !ok {
		return nil, fmt.Errorf("提议不存在: %s", proposalID)
	}

	// 计算加权投票
	totalWeightFor := 0.0
	totalWeightAgainst := 0.0
	voters := make([]string, 0)

	for _, vote := range n.votes[proposalID] {
		voters = append(voters, vote.Voter)
		switch vote.Decision {
		case VoteFor:
			totalWeightFor += vote.Weight
		case VoteAgainst:
			totalWeightAgainst += vote.Weight
		}
	}

	// 判断是否通过
	threshold := 0.5 // 需要超过50%支持
	totalWeight := totalWeightFor + totalWeightAgainst
	passed := false

	if totalWeight > 0 {
		passed = totalWeightFor/totalWeight > threshold
	} else {
		// 无投票，默认通过
		passed = true
	}

	if passed {
		proposal.Status = StatusApproved

		agreement := &Agreement{
			ID:         generateAgreementID(),
			ProposalID: proposalID,
			Parties:    voters,
			Terms:      proposal.Content,
			Timestamp:  time.Now(),
			Status:     AgreementActive,
		}

		n.agreements[agreement.ID] = agreement
		return agreement, nil
	}

	proposal.Status = StatusRejected
	return nil, fmt.Errorf("提议被拒绝")
}

// NegotiateTask 任务协商
func (n *Negotiator) NegotiateTask(ctx context.Context, task *CollabTask, agents []*AgentRole) (*TaskAssignment, error) {
	// 创建任务分配提议
	content := map[string]interface{}{
		"task_id":     task.ID,
		"task_name":   task.Name,
		"skill_id":    task.SkillID,
		"candidates":  agents,
		"priority":    task.Priority,
	}

	proposal := n.CreateProposal(task.AssignedTo, "Task Assignment: "+task.Name, ProposalTaskAssign, content)

	// 让候选Agent投票
	for _, agent := range agents {
		// 基于负载和能力评分决定投票
		weight := calculateAgentWeight(agent)
		decision := n.evaluateTaskFit(task, agent)

		n.Vote(proposal.ID, agent.AgentID, decision, "Capability evaluation", weight)
	}

	// 等待投票完成或超时
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(n.timeout):
		// 超时后解决
	}

	agreement, err := n.ResolveProposal(proposal.ID)
	if err != nil {
		return nil, err
	}

	// 从协议中提取分配
	if agentID, ok := agreement.Terms["assigned_agent"].(string); ok {
		return &TaskAssignment{
			TaskID:  task.ID,
			AgentID: agentID,
			Score:   1.0,
		}, nil
	}

	// 选择支持最多的Agent
	bestAgent := n.selectBestAgent(proposal.ID, agents)
	return &TaskAssignment{
		TaskID:  task.ID,
		AgentID: bestAgent.AgentID,
		Score:   1.0,
	}, nil
}

// evaluateTaskFit 评估任务适配度
func (n *Negotiator) evaluateTaskFit(task *CollabTask, agent *AgentRole) VoteDecision {
	// 检查能力匹配
	hasCapability := false
	for _, cap := range agent.Capabilities {
		if cap == task.SkillID {
			hasCapability = true
			break
		}
	}

	if !hasCapability {
		return VoteAgainst
	}

	// 检查负载
	if agent.CurrentLoad >= agent.MaxTasks {
		return VoteAgainst
	}

	// 根据优先级和负载决定
	if task.Priority >= agent.Priority && agent.CurrentLoad < agent.MaxTasks/2 {
		return VoteFor
	}

	return VoteNeutral
}

// selectBestAgent 选择最佳Agent
func (n *Negotiator) selectBestAgent(proposalID string, agents []*AgentRole) *AgentRole {
	n.mu.RLock()
	votes := n.votes[proposalID]
	n.mu.RUnlock()

	// 统计每个Agent的票数
	agentVotes := make(map[string]float64)
	for _, vote := range votes {
		if vote.Decision == VoteFor {
			agentVotes[vote.Voter] += vote.Weight
		}
	}

	// 选择票数最高的Agent
	var bestAgent *AgentRole
	bestScore := 0.0

	for _, agent := range agents {
		score := agentVotes[agent.AgentID]
		if score > bestScore {
			bestScore = score
			bestAgent = agent
		}
	}

	if bestAgent == nil && len(agents) > 0 {
		bestAgent = agents[0]
	}

	return bestAgent
}

// ResolveConflict 解决冲突
func (n *Negotiator) ResolveConflict(conflictType ConflictType, parties []string, data map[string]interface{}) (map[string]interface{}, error) {
	return n.conflictResolver.Resolve(conflictType, parties, data)
}

// Resolve 冲突解决
func (r *ConflictResolver) Resolve(conflictType ConflictType, parties []string, data map[string]interface{}) (map[string]interface{}, error) {
	strategy, ok := r.strategies[conflictType]
	if !ok {
		strategy = StrategyNegotiate
	}

	result := make(map[string]interface{})

	switch strategy {
	case StrategyFirstCome:
		// 第一个请求者获得
		if len(parties) > 0 {
			result["winner"] = parties[0]
		}

	case StrategyPriority:
		// 优先级最高的获得
		maxPriority := -1
		var winner string
		for _, party := range parties {
			if priority, ok := data[party+"_priority"].(int); ok {
				if priority > maxPriority {
					maxPriority = priority
					winner = party
				}
			}
		}
		result["winner"] = winner

	case StrategyVote:
		// 投票决定
		votes := make(map[string]int)
		for k, v := range data {
			if vote, ok := v.(string); ok {
				votes[vote]++
			}
		}
		var winner string
		maxVotes := 0
		for party, count := range votes {
			if count > maxVotes {
				maxVotes = count
				winner = party
			}
		}
		result["winner"] = winner

	case StrategyRandom:
		// 随机选择
		if len(parties) > 0 {
			idx := time.Now().Nanosecond() % len(parties)
			result["winner"] = parties[idx]
		}

	case StrategyNegotiate:
		// 协商解决 - 创建提议
		result["requires_negotiation"] = true
		result["parties"] = parties
	}

	return result, nil
}

// GetProposal 获取提议
func (n *Negotiator) GetProposal(proposalID string) (*Proposal, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	proposal, ok := n.proposals[proposalID]
	if !ok {
		return nil, fmt.Errorf("提议不存在: %s", proposalID)
	}
	return proposal, nil
}

// GetAgreement 获取协议
func (n *Negotiator) GetAgreement(agreementID string) (*Agreement, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	agreement, ok := n.agreements[agreementID]
	if !ok {
		return nil, fmt.Errorf("协议不存在: %s", agreementID)
	}
	return agreement, nil
}

// ListProposals 列出提议
func (n *Negotiator) ListProposals(status ProposalStatus) []*Proposal {
	n.mu.RLock()
	defer n.mu.RUnlock()

	list := make([]*Proposal, 0)
	for _, proposal := range n.proposals {
		if status == "" || proposal.Status == status {
			list = append(list, proposal)
		}
	}
	return list
}

// calculateAgentWeight 计算Agent权重
func calculateAgentWeight(agent *AgentRole) float64 {
	weight := 1.0

	// 优先级权重
	weight += float64(agent.Priority) * 0.1

	// 负载权重 (负载低权重高)
	if agent.MaxTasks > 0 {
		weight += (1.0 - float64(agent.CurrentLoad)/float64(agent.MaxTasks)) * 2.0
	}

	// 能力权重
	weight += float64(len(agent.Capabilities)) * 0.5

	return weight
}

// generateProposalID 生成提议ID
func generateProposalID() string {
	return fmt.Sprintf("prop-%d", time.Now().UnixNano())
}

// generateAgreementID 生成协议ID
func generateAgreementID() string {
	return fmt.Sprintf("agree-%d", time.Now().UnixNano())
}