package com.ofa.agent.constraint;

import java.util.ArrayList;
import java.util.List;

/**
 * 约束检查结果
 */
public class ConstraintResult {
    public final boolean allowed;
    public final ConstraintType violated;
    public final String reason;
    public final boolean requiresAuth;
    public final List<String> suggestions;

    public ConstraintResult(boolean allowed, ConstraintType violated, String reason, boolean requiresAuth) {
        this.allowed = allowed;
        this.violated = violated;
        this.reason = reason;
        this.requiresAuth = requiresAuth;
        this.suggestions = new ArrayList<>();
    }

    public void addSuggestion(String suggestion) {
        suggestions.add(suggestion);
    }
}