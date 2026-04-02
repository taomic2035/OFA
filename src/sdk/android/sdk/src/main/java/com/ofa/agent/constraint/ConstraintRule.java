package com.ofa.agent.constraint;

import java.util.regex.Pattern;

/**
 * 约束规则
 */
public class ConstraintRule {
    public final String name;
    public final ConstraintType type;
    public final Pattern actionPattern;
    public final Pattern dataPattern;
    public final boolean offlineRestricted;
    public final boolean requiresAuth;
    public final String message;

    public ConstraintRule(String name, ConstraintType type, String actionPattern, String dataPattern,
                          boolean offlineRestricted, boolean requiresAuth, String message) {
        this.name = name;
        this.type = type;
        this.actionPattern = actionPattern != null ? Pattern.compile(actionPattern, Pattern.CASE_INSENSITIVE) : null;
        this.dataPattern = dataPattern != null ? Pattern.compile(dataPattern, Pattern.CASE_INSENSITIVE) : null;
        this.offlineRestricted = offlineRestricted;
        this.requiresAuth = requiresAuth;
        this.message = message;
    }
}