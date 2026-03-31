/**
 * @file skills.hpp
 * @brief Skills Module
 * Sprint 29: C++ Agent SDK
 */

#ifndef OFA_SKILLS_HPP
#define OFA_SKILLS_HPP

#include <string>
#include <vector>
#include <map>
#include <memory>
#include <mutex>
#include "types.hpp"

namespace ofa {

/**
 * @brief 技能信息
 */
struct SkillInfo {
    std::string id;
    std::string name;
    std::string description;
    std::vector<std::string> operations;
    std::string version;
    std::string author;
    std::vector<std::string> tags;

    json toJson() const {
        return {
            {"id", id},
            {"name", name},
            {"description", description},
            {"operations", operations},
            {"version", version},
            {"author", author},
            {"tags", tags}
        };
    }
};

/**
 * @brief 技能统计
 */
struct SkillStats {
    uint64_t invocations = 0;
    uint64_t successes = 0;
    uint64_t failures = 0;
};

/**
 * @brief 技能执行器
 */
class SkillExecutor {
public:
    void registerSkill(const std::string& skillId, SkillHandler handler);
    void unregisterSkill(const std::string& skillId);

    json execute(const std::string& skillId,
                 const std::string& operation,
                 const json& input);

    std::vector<std::string> listSkills() const;
    SkillStats getStats(const std::string& skillId) const;

private:
    std::map<std::string, SkillHandler> skills_;
    std::map<std::string, SkillStats> stats_;
    mutable std::mutex mutex_;
};

/**
 * @brief 技能注册表
 */
class SkillRegistry {
public:
    void registerSkill(const SkillInfo& info);
    std::optional<SkillInfo> get(const std::string& skillId) const;
    std::vector<SkillInfo> search(const std::string& query) const;
    std::vector<SkillInfo> listAll() const;

private:
    std::map<std::string, SkillInfo> skills_;
    mutable std::mutex mutex_;
};

} // namespace ofa

#endif // OFA_SKILLS_HPP