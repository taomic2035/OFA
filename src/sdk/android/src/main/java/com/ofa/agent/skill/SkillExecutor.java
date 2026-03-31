package com.ofa.agent.skill;

/**
 * Interface for skill executors.
 * Implement this interface to create custom skills.
 */
public interface SkillExecutor {

    /**
     * Get the skill ID
     * @return unique skill identifier
     */
    String getSkillId();

    /**
     * Get skill name
     * @return human-readable skill name
     */
    String getSkillName();

    /**
     * Get skill category
     * @return category (text, data, math, utility, etc.)
     */
    String getCategory();

    /**
     * Execute the skill
     * @param input JSON-encoded input data
     * @return JSON-encoded output data
     * @throws SkillExecutionException if execution fails
     */
    byte[] execute(byte[] input) throws SkillExecutionException;
}