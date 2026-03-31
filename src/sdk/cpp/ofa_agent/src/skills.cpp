/**
 * @file skills.cpp
 * @brief Skills Implementation
 * Sprint 29: C++ Agent SDK
 */

#include "ofa/skills.hpp"
#include "ofa/error.hpp"
#include <algorithm>

namespace ofa {

void SkillExecutor::registerSkill(const std::string& skillId, SkillHandler handler) {
    std::lock_guard<std::mutex> lock(mutex_);
    skills_[skillId] = handler;
    stats_[skillId] = SkillStats{};
}

void SkillExecutor::unregisterSkill(const std::string& skillId) {
    std::lock_guard<std::mutex> lock(mutex_);
    skills_.erase(skillId);
    stats_.erase(skillId);
}

json SkillExecutor::execute(const std::string& skillId,
                            const std::string& operation,
                            const json& input) {
    std::lock_guard<std::mutex> lock(mutex_);

    auto it = skills_.find(skillId);
    if (it == skills_.end()) {
        throw SkillNotFoundException(skillId);
    }

    auto& stat = stats_[skillId];
    stat.invocations++;

    try {
        auto result = it->second(operation, input);
        stat.successes++;
        return result;
    } catch (const std::exception& e) {
        stat.failures++;
        throw;
    }
}

std::vector<std::string> SkillExecutor::listSkills() const {
    std::lock_guard<std::mutex> lock(mutex_);

    std::vector<std::string> result;
    result.reserve(skills_.size());
    for (const auto& [id, _] : skills_) {
        result.push_back(id);
    }
    return result;
}

SkillStats SkillExecutor::getStats(const std::string& skillId) const {
    std::lock_guard<std::mutex> lock(mutex_);

    auto it = stats_.find(skillId);
    if (it != stats_.end()) {
        return it->second;
    }
    return SkillStats{};
}

void SkillRegistry::registerSkill(const SkillInfo& info) {
    std::lock_guard<std::mutex> lock(mutex_);
    skills_[info.id] = info;
}

std::optional<SkillInfo> SkillRegistry::get(const std::string& skillId) const {
    std::lock_guard<std::mutex> lock(mutex_);

    auto it = skills_.find(skillId);
    if (it != skills_.end()) {
        return it->second;
    }
    return std::nullopt;
}

std::vector<SkillInfo> SkillRegistry::search(const std::string& query) const {
    std::lock_guard<std::mutex> lock(mutex_);

    std::string lowerQuery = query;
    std::transform(lowerQuery.begin(), lowerQuery.end(), lowerQuery.begin(), ::tolower);

    std::vector<SkillInfo> result;
    for (const auto& [_, info] : skills_) {
        std::string lowerName = info.name;
        std::transform(lowerName.begin(), lowerName.end(), lowerName.begin(), ::tolower);

        std::string lowerDesc = info.description;
        std::transform(lowerDesc.begin(), lowerDesc.end(), lowerDesc.begin(), ::tolower);

        if (lowerName.find(lowerQuery) != std::string::npos ||
            lowerDesc.find(lowerQuery) != std::string::npos ||
            info.id.find(lowerQuery) != std::string::npos) {
            result.push_back(info);
        }
    }
    return result;
}

std::vector<SkillInfo> SkillRegistry::listAll() const {
    std::lock_guard<std::mutex> lock(mutex_);

    std::vector<SkillInfo> result;
    result.reserve(skills_.size());
    for (const auto& [_, info] : skills_) {
        result.push_back(info);
    }
    return result;
}

} // namespace ofa