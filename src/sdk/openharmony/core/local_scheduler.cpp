/**
 * @file local_scheduler.cpp
 * @brief OFA 本地调度器实现 - 支持离线任务执行
 */

#include "ofa/agent.h"
#include <string>
#include <unordered_map>
#include <vector>
#include <queue>
#include <mutex>
#include <atomic>
#include <thread>
#include <chrono>
#include <functional>
#include <memory>

namespace ofa {

// 任务状态
enum class LocalTaskStatus {
    PENDING,
    RUNNING,
    COMPLETED,
    FAILED,
    CANCELLED
};

// 本地任务
struct LocalTask {
    std::string id;
    std::string skill_id;
    std::vector<uint8_t> input;
    std::vector<uint8_t> output;
    LocalTaskStatus status = LocalTaskStatus::PENDING;
    std::string error;
    int64_t created_at;
    int64_t completed_at;
    int retry_count = 0;
    int max_retries = 3;
};

// 工作线程
class LocalWorker {
public:
    LocalWorker(int id, std::function<void(LocalWorker*)> on_start)
        : id_(id), on_start_(on_start) {}

    void start(std::queue<std::shared_ptr<LocalTask>>* task_queue,
               std::mutex* queue_mutex,
               std::atomic<bool>* running,
               std::unordered_map<std::string, OFA_Skill>* skills,
               std::mutex* skills_mutex) {
        thread_ = std::thread([this, task_queue, queue_mutex, running, skills, skills_mutex]() {
            if (on_start_) on_start_(this);

            while (*running) {
                std::shared_ptr<LocalTask> task;

                {
                    std::lock_guard<std::mutex> lock(*queue_mutex);
                    if (!task_queue->empty()) {
                        task = task_queue->front();
                        task_queue->pop();
                    }
                }

                if (task) {
                    executeTask(task, skills, skills_mutex);
                } else {
                    std::this_thread::sleep_for(std::chrono::milliseconds(10));
                }
            }
        });
    }

    void stop() {
        if (thread_.joinable()) {
            thread_.join();
        }
    }

    int getId() const { return id_; }

private:
    void executeTask(std::shared_ptr<LocalTask> task,
                     std::unordered_map<std::string, OFA_Skill>* skills,
                     std::mutex* skills_mutex) {
        task->status = LocalTaskStatus::RUNNING;

        {
            std::lock_guard<std::mutex> lock(*skills_mutex);
            auto it = skills->find(task->skill_id);

            if (it == skills->end()) {
                task->status = LocalTaskStatus::FAILED;
                task->error = "Skill not found: " + task->skill_id;
                return;
            }

            const OFA_Skill& skill = it->second;

            if (skill.handler) {
                uint8_t* output = nullptr;
                size_t output_len = 0;

                OFA_Result result = skill.handler(
                    task->input.data(),
                    task->input.size(),
                    &output,
                    &output_len,
                    skill.user_data
                );

                if (result == OFA_OK) {
                    task->status = LocalTaskStatus::COMPLETED;
                    if (output && output_len > 0) {
                        task->output.assign(output, output + output_len);
                        OFA_Free(output);
                    }
                } else {
                    task->status = LocalTaskStatus::FAILED;
                    task->error = OFA_GetErrorString(result);

                    // 重试逻辑
                    if (task->retry_count < task->max_retries) {
                        task->retry_count++;
                        task->status = LocalTaskStatus::PENDING;
                        // TODO: 重新入队
                    }
                }
            }
        }

        task->completed_at = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()
        ).count();
    }

    int id_;
    std::thread thread_;
    std::function<void(LocalWorker*)> on_start_;
};

// 本地调度器实现
struct LocalSchedulerImpl {
    int worker_count = 4;
    OFA_OfflineLevel offline_level = OFA_OFFLINE_L1;

    std::vector<std::unique_ptr<LocalWorker>> workers;
    std::queue<std::shared_ptr<LocalTask>> task_queue;
    std::unordered_map<std::string, std::shared_ptr<LocalTask>> tasks;
    std::unordered_map<std::string, OFA_Skill> skills;

    std::mutex queue_mutex;
    std::mutex tasks_mutex;
    std::mutex skills_mutex;

    std::atomic<bool> running{false};
    std::atomic<int> pending_count{0};
    std::atomic<int> completed_count{0};

    std::function<void(const std::string&)> on_task_completed;

    LocalSchedulerImpl() = default;

    ~LocalSchedulerImpl() {
        stop();
    }

    void start() {
        if (running) return;
        running = true;

        // 创建工作线程
        for (int i = 0; i < worker_count; i++) {
            auto worker = std::make_unique<LocalWorker>(i, [](LocalWorker* w) {
                // 可选: 线程启动回调
            });
            worker->start(&task_queue, &queue_mutex, &running, &skills, &skills_mutex);
            workers.push_back(std::move(worker));
        }
    }

    void stop() {
        running = false;
        for (auto& worker : workers) {
            worker->stop();
        }
        workers.clear();
    }

    void registerSkill(const OFA_Skill& skill) {
        std::lock_guard<std::mutex> lock(skills_mutex);
        skills[skill.id] = skill;
    }

    std::string submitTask(const std::string& skill_id,
                          const std::vector<uint8_t>& input) {
        auto task = std::make_shared<LocalTask>();
        task->id = generateTaskId();
        task->skill_id = skill_id;
        task->input = input;
        task->created_at = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()
        ).count();

        {
            std::lock_guard<std::mutex> lock(tasks_mutex);
            tasks[task->id] = task;
        }

        {
            std::lock_guard<std::mutex> lock(queue_mutex);
            task_queue.push(task);
            pending_count++;
        }

        return task->id;
    }

    std::shared_ptr<LocalTask> getTask(const std::string& task_id) {
        std::lock_guard<std::mutex> lock(tasks_mutex);
        auto it = tasks.find(task_id);
        if (it != tasks.end()) {
            return it->second;
        }
        return nullptr;
    }

    bool cancelTask(const std::string& task_id) {
        std::lock_guard<std::mutex> lock(tasks_mutex);
        auto it = tasks.find(task_id);
        if (it != tasks.end()) {
            auto& task = it->second;
            if (task->status == LocalTaskStatus::PENDING) {
                task->status = LocalTaskStatus::CANCELLED;
                return true;
            }
        }
        return false;
    }

    std::vector<std::string> listPendingTasks() {
        std::vector<std::string> result;
        std::lock_guard<std::mutex> lock(tasks_mutex);
        for (const auto& pair : tasks) {
            if (pair.second->status == LocalTaskStatus::PENDING) {
                result.push_back(pair.first);
            }
        }
        return result;
    }

    std::vector<std::string> listSkills() {
        std::vector<std::string> result;
        std::lock_guard<std::mutex> lock(skills_mutex);
        for (const auto& pair : skills) {
            result.push_back(pair.first);
        }
        return result;
    }

private:
    std::string generateTaskId() {
        static std::atomic<int> counter{0};
        return "local-task-" + std::to_string(counter++);
    }
};

} // namespace ofa

// C 接口实现

struct OFA_LocalScheduler {
    ofa::LocalSchedulerImpl impl;
};

OFA_LocalScheduler* OFA_LocalScheduler_Create(int worker_count, OFA_OfflineLevel level) {
    auto* scheduler = new OFA_LocalScheduler();
    scheduler->impl.worker_count = worker_count > 0 ? worker_count : 4;
    scheduler->impl.offline_level = level;
    return scheduler;
}

OFA_Result OFA_LocalScheduler_Start(OFA_LocalScheduler* scheduler) {
    if (!scheduler) return OFA_ERROR;
    scheduler->impl.start();
    return OFA_OK;
}

OFA_Result OFA_LocalScheduler_Stop(OFA_LocalScheduler* scheduler) {
    if (!scheduler) return OFA_ERROR;
    scheduler->impl.stop();
    return OFA_OK;
}

void OFA_LocalScheduler_Destroy(OFA_LocalScheduler* scheduler) {
    if (scheduler) {
        scheduler->impl.stop();
        delete scheduler;
    }
}

OFA_Result OFA_LocalScheduler_RegisterSkill(
    OFA_LocalScheduler* scheduler,
    const OFA_Skill* skill
) {
    if (!scheduler || !skill) return OFA_ERROR;
    scheduler->impl.registerSkill(*skill);
    return OFA_OK;
}

OFA_Result OFA_LocalScheduler_SubmitTask(
    OFA_LocalScheduler* scheduler,
    const char* skill_id,
    const uint8_t* input,
    size_t input_len,
    char** task_id
) {
    if (!scheduler || !skill_id) return OFA_ERROR;

    std::vector<uint8_t> input_data;
    if (input && input_len > 0) {
        input_data.assign(input, input + input_len);
    }

    std::string id = scheduler->impl.submitTask(skill_id, input_data);
    if (task_id) {
        *task_id = strdup(id.c_str());
    }

    return OFA_OK;
}

OFA_Result OFA_LocalScheduler_GetTaskStatus(
    OFA_LocalScheduler* scheduler,
    const char* task_id,
    OFA_LocalTaskStatus* status,
    const uint8_t** output,
    size_t* output_len
) {
    if (!scheduler || !task_id) return OFA_ERROR;

    auto task = scheduler->impl.getTask(task_id);
    if (!task) return OFA_ERROR;

    if (status) {
        switch (task->status) {
            case ofa::LocalTaskStatus::PENDING:
                *status = OFA_LOCAL_TASK_PENDING;
                break;
            case ofa::LocalTaskStatus::RUNNING:
                *status = OFA_LOCAL_TASK_RUNNING;
                break;
            case ofa::LocalTaskStatus::COMPLETED:
                *status = OFA_LOCAL_TASK_COMPLETED;
                break;
            case ofa::LocalTaskStatus::FAILED:
                *status = OFA_LOCAL_TASK_FAILED;
                break;
            case ofa::LocalTaskStatus::CANCELLED:
                *status = OFA_LOCAL_TASK_CANCELLED;
                break;
        }
    }

    if (output && output_len) {
        *output = task->output.data();
        *output_len = task->output.size();
    }

    return OFA_OK;
}

OFA_Result OFA_LocalScheduler_CancelTask(
    OFA_LocalScheduler* scheduler,
    const char* task_id
) {
    if (!scheduler || !task_id) return OFA_ERROR;

    if (scheduler->impl.cancelTask(task_id)) {
        return OFA_OK;
    }
    return OFA_ERROR;
}

size_t OFA_LocalScheduler_GetPendingCount(const OFA_LocalScheduler* scheduler) {
    if (!scheduler) return 0;
    return scheduler->impl.pending_count;
}

size_t OFA_LocalScheduler_GetCompletedCount(const OFA_LocalScheduler* scheduler) {
    if (!scheduler) return 0;
    return scheduler->impl.completed_count;
}