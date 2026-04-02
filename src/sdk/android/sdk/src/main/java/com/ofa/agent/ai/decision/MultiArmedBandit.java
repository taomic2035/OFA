package com.ofa.agent.ai.decision;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Random;

/**
 * Multi-Armed Bandit - implements bandit algorithms for intelligent action selection.
 * Supports Epsilon-Greedy, UCB, and Thompson Sampling strategies.
 */
public class MultiArmedBandit {

    private static final String TAG = "MultiArmedBandit";

    private final String id;
    private final Strategy strategy;
    private final Map<String, Arm> arms;
    private final Random random;

    private double epsilon = 0.1; // For epsilon-greedy
    private double explorationBonus = 2.0; // For UCB

    /**
     * Selection strategy
     */
    public enum Strategy {
        EPSILON_GREEDY,  // Simple exploration
        UCB,             // Upper Confidence Bound
        THOMPSON_SAMPLING // Thompson Sampling
    }

    /**
     * Arm statistics
     */
    private static class Arm {
        final String id;
        int pulls = 0;
        double totalReward = 0;
        double sumSquaredReward = 0;

        // For Thompson Sampling (Beta distribution parameters)
        double alpha = 1.0;
        double beta = 1.0;

        Arm(String id) {
            this.id = id;
        }

        double getAverageReward() {
            return pulls > 0 ? totalReward / pulls : 0;
        }

        double getVariance() {
            if (pulls < 2) return 0;
            double mean = getAverageReward();
            return sumSquaredReward / pulls - mean * mean;
        }

        void update(double reward) {
            pulls++;
            totalReward += reward;
            sumSquaredReward += reward * reward;

            // Update Beta parameters for Thompson Sampling
            if (reward > 0.5) {
                alpha += reward;
            } else {
                beta += (1 - reward);
            }
        }
    }

    /**
     * Create multi-armed bandit
     */
    public MultiArmedBandit(@NonNull String id, @NonNull Strategy strategy) {
        this.id = id;
        this.strategy = strategy;
        this.arms = new HashMap<>();
        this.random = new Random();
    }

    /**
     * Get bandit ID
     */
    @NonNull
    public String getId() {
        return id;
    }

    /**
     * Add an arm
     */
    public void addArm(@NonNull String armId) {
        arms.put(armId, new Arm(armId));
        Log.d(TAG, "Added arm: " + armId + " to bandit: " + id);
    }

    /**
     * Remove an arm
     */
    public void removeArm(@NonNull String armId) {
        arms.remove(armId);
    }

    /**
     * Select an arm based on strategy
     */
    @Nullable
    public String selectArm() {
        if (arms.isEmpty()) {
            return null;
        }

        // Initialize arms with no pulls
        for (Arm arm : arms.values()) {
            if (arm.pulls == 0) {
                Log.d(TAG, "Selecting unexplored arm: " + arm.id);
                return arm.id;
            }
        }

        switch (strategy) {
            case EPSILON_GREEDY:
                return selectEpsilonGreedy();
            case UCB:
                return selectUCB();
            case THOMPSON_SAMPLING:
                return selectThompsonSampling();
            default:
                return selectEpsilonGreedy();
        }
    }

    /**
     * Epsilon-Greedy selection
     */
    @NonNull
    private String selectEpsilonGreedy() {
        // Explore with probability epsilon
        if (random.nextDouble() < epsilon) {
            List<String> armIds = new ArrayList<>(arms.keySet());
            String selected = armIds.get(random.nextInt(armIds.size()));
            Log.d(TAG, "Exploration: selected " + selected);
            return selected;
        }

        // Exploit: select best arm
        String bestArm = null;
        double bestValue = Double.NEGATIVE_INFINITY;

        for (Arm arm : arms.values()) {
            double value = arm.getAverageReward();
            if (value > bestValue) {
                bestValue = value;
                bestArm = arm.id;
            }
        }

        Log.d(TAG, "Exploitation: selected " + bestArm + " (value=" + bestValue + ")");
        return bestArm;
    }

    /**
     * UCB (Upper Confidence Bound) selection
     */
    @NonNull
    private String selectUCB() {
        String bestArm = null;
        double bestUcb = Double.NEGATIVE_INFINITY;

        int totalPulls = getTotalPulls();

        for (Arm arm : arms.values()) {
            double ucb = arm.getAverageReward() +
                explorationBonus * Math.sqrt(Math.log(totalPulls) / arm.pulls);

            if (ucb > bestUcb) {
                bestUcb = ucb;
                bestArm = arm.id;
            }
        }

        Log.d(TAG, "UCB: selected " + bestArm + " (ucb=" + bestUcb + ")");
        return bestArm;
    }

    /**
     * Thompson Sampling selection
     */
    @NonNull
    private String selectThompsonSampling() {
        String bestArm = null;
        double bestSample = Double.NEGATIVE_INFINITY;

        for (Arm arm : arms.values()) {
            // Sample from Beta distribution
            double sample = sampleBeta(arm.alpha, arm.beta);

            if (sample > bestSample) {
                bestSample = sample;
                bestArm = arm.id;
            }
        }

        Log.d(TAG, "Thompson Sampling: selected " + bestArm + " (sample=" + bestSample + ")");
        return bestArm;
    }

    /**
     * Sample from Beta distribution using Gamma distribution
     */
    private double sampleBeta(double alpha, double beta) {
        double x = sampleGamma(alpha);
        double y = sampleGamma(beta);
        return x / (x + y);
    }

    /**
     * Sample from Gamma distribution (simplified)
     */
    private double sampleGamma(double shape) {
        if (shape < 1) {
            return sampleGamma(shape + 1) * Math.pow(random.nextDouble(), 1.0 / shape);
        }

        double d = shape - 1.0 / 3.0;
        double c = 1.0 / Math.sqrt(9.0 * d);

        while (true) {
            double x = random.nextGaussian();
            double v = 1.0 + c * x;

            if (v <= 0) continue;

            v = v * v * v;
            double u = random.nextDouble();

            if (u < 1.0 - 0.0331 * (x * x) * (x * x)) {
                return d * v;
            }

            if (Math.log(u) < 0.5 * x * x + d * (1.0 - v + Math.log(v))) {
                return d * v;
            }
        }
    }

    /**
     * Update arm with reward
     */
    public void update(@NonNull String armId, double reward) {
        Arm arm = arms.get(armId);
        if (arm == null) {
            Log.w(TAG, "Arm not found: " + armId);
            return;
        }

        // Clamp reward to [0, 1]
        reward = Math.max(0, Math.min(1, reward));
        arm.update(reward);

        Log.d(TAG, "Updated arm " + armId + " with reward " + reward +
            " (avg=" + arm.getAverageReward() + ", pulls=" + arm.pulls + ")");
    }

    /**
     * Get total pulls across all arms
     */
    public int getTotalPulls() {
        int total = 0;
        for (Arm arm : arms.values()) {
            total += arm.pulls;
        }
        return total;
    }

    /**
     * Get arm statistics
     */
    @NonNull
    public Map<String, Double> getArmScores() {
        Map<String, Double> scores = new HashMap<>();
        for (Map.Entry<String, Arm> entry : arms.entrySet()) {
            scores.put(entry.getKey(), entry.getValue().getAverageReward());
        }
        return scores;
    }

    /**
     * Get best arm
     */
    @Nullable
    public String getBestArm() {
        String bestArm = null;
        double bestValue = Double.NEGATIVE_INFINITY;

        for (Arm arm : arms.values()) {
            if (arm.pulls > 0 && arm.getAverageReward() > bestValue) {
                bestValue = arm.getAverageReward();
                bestArm = arm.id;
            }
        }

        return bestArm;
    }

    /**
     * Set epsilon for epsilon-greedy
     */
    public void setEpsilon(double epsilon) {
        this.epsilon = Math.max(0, Math.min(1, epsilon));
    }

    /**
     * Set exploration bonus for UCB
     */
    public void setExplorationBonus(double bonus) {
        this.explorationBonus = bonus;
    }

    /**
     * Reset all arm statistics
     */
    public void reset() {
        for (Arm arm : arms.values()) {
            arm.pulls = 0;
            arm.totalReward = 0;
            arm.sumSquaredReward = 0;
            arm.alpha = 1.0;
            arm.beta = 1.0;
        }
        Log.i(TAG, "Reset bandit: " + id);
    }

    /**
     * Get statistics as string
     */
    @NonNull
    public String getStats() {
        StringBuilder sb = new StringBuilder();
        sb.append("Bandit{id=").append(id);
        sb.append(", strategy=").append(strategy);
        sb.append(", arms=").append(arms.size());
        sb.append(", totalPulls=").append(getTotalPulls());
        sb.append(", bestArm=").append(getBestArm());
        sb.append("}");
        return sb.toString();
    }
}