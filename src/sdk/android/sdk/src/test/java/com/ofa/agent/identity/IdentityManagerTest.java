package com.ofa.agent.identity;

import android.content.Context;

import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.robolectric.RobolectricTestRunner;
import org.robolectric.RuntimeEnvironment;
import org.robolectric.annotation.Config;

import java.util.HashMap;
import java.util.Map;

import static org.junit.Assert.*;

/**
 * Unit tests for IdentityManager
 */
@RunWith(RobolectricTestRunner.class)
@Config(sdk = 28)
public class IdentityManagerTest {

    private Context context;
    private IdentityManager identityManager;

    @Before
    public void setUp() {
        context = RuntimeEnvironment.application;
        identityManager = new IdentityManager(context);
        identityManager.initialize();
    }

    @Test
    public void testInitialization() {
        assertTrue("IdentityManager should be initialized", identityManager.isInitialized());
        assertTrue("IdentityManager should have identity", identityManager.hasIdentity());
    }

    @Test
    public void testCreateIdentity() {
        PersonalIdentity identity = identityManager.getIdentity();
        assertNotNull("Identity should not be null", identity);
        assertNotNull("Identity ID should not be null", identity.getId());
        assertTrue("Identity ID should not be empty", !identity.getId().isEmpty());
    }

    @Test
    public void testUpdateBasicInfo() {
        String testName = "Test User";
        String testNickname = "TestNick";
        String testAvatar = "https://example.com/avatar.png";
        String testLocation = "Beijing";

        identityManager.updateBasicInfo(testName, testNickname, testAvatar, testLocation);

        PersonalIdentity identity = identityManager.getIdentity();
        assertEquals("Name should be updated", testName, identity.getName());
        assertEquals("Nickname should be updated", testNickname, identity.getNickname());
        assertEquals("Avatar should be updated", testAvatar, identity.getAvatar());
        assertEquals("Location should be updated", testLocation, identity.getLocation());
    }

    @Test
    public void testUpdatePersonality() {
        Map<String, Double> updates = new HashMap<>();
        updates.put("openness", 0.8);
        updates.put("conscientiousness", 0.7);
        updates.put("extraversion", 0.6);

        identityManager.updatePersonality(updates);

        PersonalIdentity identity = identityManager.getIdentity();
        Personality personality = identity.getPersonality();
        assertNotNull("Personality should not be null", personality);
        assertEquals("Openness should be updated", 0.8, personality.getOpenness(), 0.01);
        assertEquals("Conscientiousness should be updated", 0.7, personality.getConscientiousness(), 0.01);
        assertEquals("Extraversion should be updated", 0.6, personality.getExtraversion(), 0.01);
    }

    @Test
    public void testSetMBTIType() {
        identityManager.setMBTIType("INTJ");

        PersonalIdentity identity = identityManager.getIdentity();
        Personality personality = identity.getPersonality();
        assertEquals("MBTI type should be INTJ", "INTJ", personality.getMbtiType());
        assertTrue("MBTI confidence should be high", personality.getMbtiConfidence() > 0.8);
    }

    @Test
    public void testUpdateValueSystem() {
        Map<String, Double> updates = new HashMap<>();
        updates.put("privacy", 0.9);
        updates.put("efficiency", 0.8);
        updates.put("health", 0.7);

        identityManager.updateValueSystem(updates);

        PersonalIdentity identity = identityManager.getIdentity();
        ValueSystem valueSystem = identity.getValueSystem();
        assertNotNull("ValueSystem should not be null", valueSystem);
        assertEquals("Privacy should be updated", 0.9, valueSystem.getPrivacy(), 0.01);
        assertEquals("Efficiency should be updated", 0.8, valueSystem.getEfficiency(), 0.01);
        assertEquals("Health should be updated", 0.7, valueSystem.getHealth(), 0.01);
    }

    @Test
    public void testAddInterest() {
        Interest interest = new Interest(Interest.CATEGORY_TECH, "Programming");
        interest.addKeyword("Java");
        interest.addKeyword("Android");

        identityManager.addInterest(interest);

        PersonalIdentity identity = identityManager.getIdentity();
        assertTrue("Should have interests", identity.getInterests().size() > 0);
        Interest added = identity.getInterests().get(0);
        assertEquals("Interest category should match", Interest.CATEGORY_TECH, added.getCategory());
        assertEquals("Interest name should match", "Programming", added.getName());
        assertTrue("Should have keywords", added.getKeywords().size() >= 2);
    }

    @Test
    public void testRemoveInterest() {
        Interest interest = new Interest(Interest.CATEGORY_MUSIC, "Guitar");
        identityManager.addInterest(interest);

        PersonalIdentity identity = identityManager.getIdentity();
        String interestId = identity.getInterests().get(0).getId();

        assertTrue("Remove should succeed", identityManager.removeInterest(interestId));
        identity = identityManager.getIdentity();
        for (Interest i : identity.getInterests()) {
            assertNotEquals("Interest should be removed", interestId, i.getId());
        }
    }

    @Test
    public void testDecisionContext() {
        DecisionContext context = identityManager.getDecisionContext();
        assertNotNull("DecisionContext should not be null", context);
        assertEquals("Identity ID should match", identityManager.getIdentityId(), context.getIdentityId());
    }

    @Test
    public void testGeneratePromptContext() {
        identityManager.updateBasicInfo("Test User", null, null, null);

        String promptContext = identityManager.generatePromptContext();
        assertNotNull("Prompt context should not be null", promptContext);
        assertTrue("Prompt context should contain user context", promptContext.contains("用户身份上下文"));
    }

    @Test
    public void testObserveBehavior() {
        Map<String, Object> context = new HashMap<>();
        context.put("decision_type", "impulse_purchase");

        identityManager.observeBehavior(BehaviorObservation.TYPE_DECISION, context);

        // Verify behavior is processed (indirectly through personality update)
        // After enough observations, personality should change
    }

    @Test
    public void testPersonalityDescription() {
        Map<String, Double> updates = new HashMap<>();
        updates.put("openness", 0.8);
        updates.put("extraversion", 0.3);

        identityManager.updatePersonality(updates);

        PersonalIdentity identity = identityManager.getIdentity();
        String desc = identity.getPersonalityDescription();
        assertNotNull("Description should not be null", desc);
        assertTrue("Description should mention creativity", desc.contains("创造力") || desc.contains("好奇"));
        assertTrue("Description should mention introversion", desc.contains("内向"));
    }

    @Test
    public void testValuePriority() {
        Map<String, Double> updates = new HashMap<>();
        updates.put("privacy", 0.9);
        updates.put("family", 0.8);
        updates.put("career", 0.6);

        identityManager.updateValueSystem(updates);

        PersonalIdentity identity = identityManager.getIdentity();
        String[] priority = identity.getValuePriority();
        assertTrue("Should have priorities", priority.length > 0);
        assertEquals("Privacy should be top priority", "privacy", priority[0]);
    }

    @Test
    public void testVersionIncrement() {
        PersonalIdentity identity = identityManager.getIdentity();
        long initialVersion = identity.getVersion();

        identityManager.updateBasicInfo("New Name", null, null, null);

        identity = identityManager.getIdentity();
        assertTrue("Version should increment", identity.getVersion() > initialVersion);
    }

    @Test
    public void testShutdown() {
        identityManager.shutdown();
        assertFalse("Should not be initialized after shutdown", identityManager.isInitialized());
    }
}