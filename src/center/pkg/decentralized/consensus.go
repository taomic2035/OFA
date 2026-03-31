// Package decentralized - 共识引擎
package decentralized

import (
	"fmt"
	"sync"
	"time"
)

// ConsensusEngine 共识引擎
type ConsensusEngine struct {
	algorithm    ConsensusAlgorithm
	proposals    map[string]*Proposal
	votes        map[string][]*Vote
	validators   map[string]*ValidatorInfo
	threshold    float64 // 通过阈值
	mu           sync.RWMutex
}

// ConsensusAlgorithm 共识算法
type ConsensusAlgorithm string

const (
	AlgorithmPBFT      ConsensusAlgorithm = "pbft"      // PBFT
	AlgorithmRaft      ConsensusAlgorithm = "raft"      // Raft
	AlgorithmPOA       ConsensusAlgorithm = "poa"       // Proof of Authority
	AlgorithmPOS       ConsensusAlgorithm = "pos"       // Proof of Stake
	AlgorithmVoting    ConsensusAlgorithm = "voting"    // 简单投票
	AlgorithmWeighted  ConsensusAlgorithm = "weighted"  // 加权投票
)

// ValidatorInfo 验证者信息
type ValidatorInfo struct {
	ID       string    `json:"id"`
	Weight   float64   `json:"weight"`
	Status   string    `json:"status"`
	LastVote time.Time `json:"last_vote"`
}

// Proposal 提案
type Proposal struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Subject     string                 `json:"subject"`
	Content     map[string]interface{} `json:"content"`
	Proposer    string                 `json:"proposer"`
	Status      ProposalStatus         `json:"status"`
	Votes       []*Vote                `json:"votes"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	VotesFor    int                    `json:"votes_for"`
	VotesAgainst int                   `json:"votes_against"`
}

// ProposalStatus 提案状态
type ProposalStatus string

const (
	ProposalStatusPending  ProposalStatus = "pending"
	ProposalStatusVoting   ProposalStatus = "voting"
	ProposalStatusApproved ProposalStatus = "approved"
	ProposalStatusRejected ProposalStatus = "rejected"
	ProposalStatusExpired  ProposalStatus = "expired"
)

// Vote 投票
type Vote struct {
	VoterID    string    `json:"voter_id"`
	ProposalID string    `json:"proposal_id"`
	Decision   VoteDecision `json:"decision"`
	Weight     float64   `json:"weight"`
	Timestamp  time.Time `json:"timestamp"`
	Reason     string    `json:"reason"`
}

// VoteDecision 投票决策
type VoteDecision string

const (
	VoteApprove VoteDecision = "approve"
	VoteReject  VoteDecision = "reject"
	VoteAbstain VoteDecision = "abstain"
)

// ConsensusDecision 共识决策
type ConsensusDecision struct {
	ProposalID   string                 `json:"proposal_id"`
	Approved     bool                   `json:"approved"`
	SelectedNode string                 `json:"selected_node,omitempty"`
	Terms        map[string]interface{} `json:"terms,omitempty"`
	VotesFor     int                    `json:"votes_for"`
	VotesAgainst int                    `json:"votes_against"`
	Confidence   float64                `json:"confidence"`
	Timestamp    time.Time              `json:"timestamp"`
}

// NewConsensusEngine 创建共识引擎
func NewConsensusEngine() *ConsensusEngine {
	return &ConsensusEngine{
		algorithm:   AlgorithmWeighted,
		proposals:   make(map[string]*Proposal),
		votes:       make(map[string][]*Vote),
		validators:  make(map[string]*ValidatorInfo),
		threshold:   0.6, // 60%阈值
	}
}

// SetAlgorithm 设置共识算法
func (e *ConsensusEngine) SetAlgorithm(algorithm ConsensusAlgorithm) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.algorithm = algorithm
}

// SetThreshold 设置阈值
func (e *ConsensusEngine) SetThreshold(threshold float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.threshold = threshold
}

// AddValidator 添加验证者
func (e *ConsensusEngine) AddValidator(id string, weight float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.validators[id] = &ValidatorInfo{
		ID:     id,
		Weight: weight,
		Status: "active",
	}
}

// RemoveValidator 移除验证者
func (e *ConsensusEngine) RemoveValidator(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.validators, id)
}

// CreateProposal 创建提案
func (e *ConsensusEngine) CreateProposal(typeStr, subject string, content interface{}, proposer string) *Proposal {
	e.mu.Lock()
	defer e.mu.Unlock()

	proposal := &Proposal{
		ID:        generateProposalID(),
		Type:      typeStr,
		Subject:   subject,
		Proposer:  proposer,
		Status:    ProposalStatusVoting,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Votes:     make([]*Vote, 0),
	}

	// 转换content
	if c, ok := content.(map[string]interface{}); ok {
		proposal.Content = c
	}

	e.proposals[proposal.ID] = proposal
	e.votes[proposal.ID] = make([]*Vote, 0)

	return proposal
}

// EvaluateProposal 评估提案
func (e *ConsensusEngine) EvaluateProposal(proposal *Proposal, node *NodeInfo) *Vote {
	e.mu.RLock()
	weight := 1.0
	if validator, ok := e.validators[node.ID]; ok {
		weight = validator.Weight
	}
	e.mu.RUnlock()

	// 基于节点状态和信任评分决策
	decision := e.makeDecision(proposal, node)

	return &Vote{
		VoterID:    node.ID,
		ProposalID: proposal.ID,
		Decision:   decision,
		Weight:     weight,
		Timestamp:  time.Now(),
	}
}

// makeDecision 做出决策
func (e *ConsensusEngine) makeDecision(proposal *Proposal, node *NodeInfo) VoteDecision {
	// 基于信任评分
	if node.TrustScore < 0.3 {
		return VoteAbstain
	}

	// 基于提案类型
	switch proposal.Type {
	case "task_execution":
		// 检查任务能力
		if node.Capabilities != nil && len(node.Capabilities) > 0 {
			return VoteApprove
		}
		return VoteReject

	case "resource_allocation":
		// 检查负载
		if node.Load < node.MaxLoad {
			return VoteApprove
		}
		return VoteReject

	case "policy_update":
		// 高信任评分自动同意
		if node.TrustScore > 0.7 {
			return VoteApprove
		}
		return VoteAbstain

	default:
		return VoteApprove
	}
}

// MakeDecision 做出共识决策
func (e *ConsensusEngine) MakeDecision(proposal *Proposal) (*ConsensusDecision, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 检查提案状态
	if proposal.Status != ProposalStatusVoting {
		return nil, fmt.Errorf("提案不在投票状态")
	}

	// 检查是否过期
	if time.Now().After(proposal.ExpiresAt) {
		proposal.Status = ProposalStatusExpired
		return nil, fmt.Errorf("提案已过期")
	}

	// 计算投票权重
	totalWeight := 0.0
	approveWeight := 0.0
	rejectWeight := 0.0

	for _, vote := range proposal.Votes {
		totalWeight += vote.Weight
		switch vote.Decision {
		case VoteApprove:
			approveWeight += vote.Weight
		case VoteReject:
			rejectWeight += vote.Weight
		}
	}

	decision := &ConsensusDecision{
		ProposalID:   proposal.ID,
		VotesFor:     proposal.VotesFor,
		VotesAgainst: proposal.VotesAgainst,
		Timestamp:    time.Now(),
	}

	// 判断是否通过
	if totalWeight > 0 {
		ratio := approveWeight / totalWeight
		decision.Confidence = ratio

		if ratio >= e.threshold {
			decision.Approved = true
			proposal.Status = ProposalStatusApproved

			// 提取选定节点
			if selected, ok := proposal.Content["selected_node"].(string); ok {
				decision.SelectedNode = selected
			}
		} else {
			decision.Approved = false
			proposal.Status = ProposalStatusRejected
		}
	} else {
		// 无投票，默认通过
		decision.Approved = true
		decision.Confidence = 1.0
		proposal.Status = ProposalStatusApproved
	}

	// 复制条款
	decision.Terms = proposal.Content

	return decision, nil
}

// SubmitVote 提交投票
func (e *ConsensusEngine) SubmitVote(vote *Vote) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	proposal, ok := e.proposals[vote.ProposalID]
	if !ok {
		return fmt.Errorf("提案不存在: %s", vote.ProposalID)
	}

	// 检查是否已投票
	for _, v := range proposal.Votes {
		if v.VoterID == vote.VoterID {
			return fmt.Errorf("已投票: %s", vote.VoterID)
		}
	}

	// 添加投票
	proposal.Votes = append(proposal.Votes, vote)

	switch vote.Decision {
	case VoteApprove:
		proposal.VotesFor++
	case VoteReject:
		proposal.VotesAgainst++
	}

	return nil
}

// GetProposal 获取提案
func (e *ConsensusEngine) GetProposal(proposalID string) (*Proposal, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	proposal, ok := e.proposals[proposalID]
	if !ok {
		return nil, fmt.Errorf("提案不存在: %s", proposalID)
	}
	return proposal, nil
}

// ListProposals 列出提案
func (e *ConsensusEngine) ListProposals(status ProposalStatus) []*Proposal {
	e.mu.RLock()
	defer e.mu.RUnlock()

	list := make([]*Proposal, 0)
	for _, proposal := range e.proposals {
		if status == "" || proposal.Status == status {
			list = append(list, proposal)
		}
	}
	return list
}

// generateProposalID 生成提案ID
func generateProposalID() string {
	return fmt.Sprintf("prop-%d", time.Now().UnixNano())
}

// ConsensusStats 共识统计
type ConsensusStats struct {
	TotalProposals    int     `json:"total_proposals"`
	ApprovedProposals int     `json:"approved_proposals"`
	RejectedProposals int     `json:"rejected_proposals"`
	ExpiredProposals  int     `json:"expired_proposals"`
	TotalValidators   int     `json:"total_validators"`
	AvgConfidence     float64 `json:"avg_confidence"`
	AvgDurationMs     int64   `json:"avg_duration_ms"`
}

// GetStats 获取统计
func (e *ConsensusEngine) GetStats() *ConsensusStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stats := &ConsensusStats{
		TotalProposals:  len(e.proposals),
		TotalValidators: len(e.validators),
	}

	if len(e.proposals) == 0 {
		return stats
	}

	totalConfidence := 0.0
	totalDuration := 0.0

	for _, proposal := range e.proposals {
		switch proposal.Status {
		case ProposalStatusApproved:
			stats.ApprovedProposals++
		case ProposalStatusRejected:
			stats.RejectedProposals++
		case ProposalStatusExpired:
			stats.ExpiredProposals++
		}

		duration := time.Since(proposal.CreatedAt).Milliseconds()
		totalDuration += float64(duration)
	}

	stats.AvgConfidence = totalConfidence / float64(len(e.proposals))
	stats.AvgDurationMs = int64(totalDuration / float64(len(e.proposals)))

	return stats
}