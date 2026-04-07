package com.ofa.agent.behavior;

import static org.junit.Assert.*;

import org.junit.Before;
import org.junit.Test;

import java.util.HashMap;
import java.util.Map;

/**
 * BehaviorCollector 单元测试
 */
public class BehaviorCollectorTest {

    private TestIdentityManager identityManager;
    private BehaviorCollector collector;

    @Before
    public void setUp() {
        identityManager = new TestIdentityManager();
        // 注意：实际使用需要 Context，这里用 mock
    }

    @Test
    public void testObserveDecision() {
        Map<String, Object> details = new HashMap<>();
        details.put("item", "coffee");
        details.put("price", 25.0);

        BehaviorObservation obs = new BehaviorObservation(
            BehaviorObservation.TYPE_DECISION, details);

        // 触发自动推断
        obs.autoInfer();

        assertNotNull(obs.getId());
        assertEquals(BehaviorObservation.TYPE_DECISION, obs.getType());
        assertNotNull(obs.getInferences());
    }

    @Test
    public void testImpulsePurchaseInference() {
        Map<String, Object> details = new HashMap<>();
        details.put("decision_type", "impulse_purchase");

        BehaviorObservation obs = new BehaviorObservation(
            BehaviorObservation.TYPE_DECISION, details);
        obs.autoInfer();

        Map<String, Double> inferences = obs.getInferences();

        // 冲动购买应该增加神经质
        assertTrue(inferences.containsKey("neuroticism"));
        assertTrue(inferences.get("neuroticism") > 0);

        // 冲动购买应该降低尽责性
        assertTrue(inferences.containsKey("conscientiousness"));
        assertTrue(inferences.get("conscientiousness") < 0);
    }

    @Test
    public void testGroupChatsInference() {
        Map<String, Object> details = new HashMap<>();
        details.put("interaction_type", "group_chats");

        BehaviorObservation obs = new BehaviorObservation(
            BehaviorObservation.TYPE_INTERACTION, details);
        obs.autoInfer();

        Map<String, Double> inferences = obs.getInferences();

        // 群聊应该增加外向性
        assertTrue(inferences.containsKey("extraversion"));
        assertTrue(inferences.get("extraversion") > 0);
    }

    @Test
    public void testNovelTryingInference() {
        Map<String, Object> details = new HashMap<>();
        details.put("preference_type", "novel_trying");

        BehaviorObservation obs = new BehaviorObservation(
            BehaviorObservation.TYPE_PREFERENCE, details);
        obs.autoInfer();

        Map<String, Double> inferences = obs.getInferences();

        // 尝试新事物应该增加开放性
        assertTrue(inferences.containsKey("openness"));
        assertTrue(inferences.get("openness") > 0);
    }

    @Test
    public void testRegularScheduleInference() {
        Map<String, Object> details = new HashMap<>();
        details.put("activity_type", "regular_schedule");

        BehaviorObservation obs = new BehaviorObservation(
            BehaviorObservation.TYPE_ACTIVITY, details);
        obs.autoInfer();

        Map<String, Double> inferences = obs.getInferences();

        // 规律作息应该增加尽责性
        assertTrue(inferences.containsKey("conscientiousness"));
        assertTrue(inferences.get("conscientiousness") > 0);
    }

    @Test
    public void testBehaviorObservationToJson() {
        Map<String, Object> details = new HashMap<>();
        details.put("test_key", "test_value");

        BehaviorObservation obs = new BehaviorObservation(
            BehaviorObservation.TYPE_DECISION, details);

        String json = obs.toJson();

        assertNotNull(json);
        assertTrue(json.contains("\"type\":\"decision\""));
    }

    /**
     * 测试用 IdentityManager
     */
    private static class TestIdentityManager {
        // 简化的测试用 IdentityManager
    }
}