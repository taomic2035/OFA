package com.ofa.agent.skill;

/**
 * Exception thrown when skill execution fails
 */
public class SkillExecutionException extends Exception {

    private final String skillId;
    private final ErrorCode errorCode;

    public enum ErrorCode {
        INVALID_INPUT(400, "Invalid input"),
        SKILL_NOT_FOUND(404, "Skill not found"),
        EXECUTION_ERROR(500, "Execution error"),
        TIMEOUT(504, "Timeout");

        private final int code;
        private final String message;

        ErrorCode(int code, String message) {
            this.code = code;
            this.message = message;
        }

        public int getCode() {
            return code;
        }

        public String getMessage() {
            return message;
        }
    }

    public SkillExecutionException(String skillId, String message) {
        super(message);
        this.skillId = skillId;
        this.errorCode = ErrorCode.EXECUTION_ERROR;
    }

    public SkillExecutionException(String skillId, ErrorCode errorCode, String message) {
        super(message);
        this.skillId = skillId;
        this.errorCode = errorCode;
    }

    public SkillExecutionException(String skillId, String message, Throwable cause) {
        super(message, cause);
        this.skillId = skillId;
        this.errorCode = ErrorCode.EXECUTION_ERROR;
    }

    public String getSkillId() {
        return skillId;
    }

    public ErrorCode getErrorCode() {
        return errorCode;
    }
}