package com.ofa.agent.skill.builtin;

import org.junit.Test;
import org.junit.Before;

import static org.junit.Assert.*;

/**
 * Unit tests for built-in offline skills
 */
public class OfflineSkillsTest {

    private CalculatorSkill calculator;
    private TimestampSkill timestamp;
    private JSONFormatSkill jsonFormat;
    private HashSkill hash;
    private EchoSkill echo;

    @Before
    public void setUp() {
        calculator = new CalculatorSkill();
        timestamp = new TimestampSkill();
        jsonFormat = new JSONFormatSkill();
        hash = new HashSkill();
        echo = new EchoSkill();
    }

    // ===== Calculator Tests =====

    @Test
    public void testCalculatorAddition() throws Exception {
        byte[] result = calculator.execute("5 + 3".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"result\":8.0") || json.contains("\"result\":8"));
    }

    @Test
    public void testCalculatorSubtraction() throws Exception {
        byte[] result = calculator.execute("10 - 4".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"result\":6.0") || json.contains("\"result\":6"));
    }

    @Test
    public void testCalculatorMultiplication() throws Exception {
        byte[] result = calculator.execute("7 * 6".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"result\":42.0") || json.contains("\"result\":42"));
    }

    @Test
    public void testCalculatorDivision() throws Exception {
        byte[] result = calculator.execute("20 / 4".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"result\":5.0") || json.contains("\"result\":5"));
    }

    @Test
    public void testCalculatorSqrt() throws Exception {
        byte[] result = calculator.execute("sqrt 16".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"result\":4.0") || json.contains("\"result\":4"));
    }

    @Test
    public void testCalculatorPower() throws Exception {
        byte[] result = calculator.execute("2 pow 10".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"result\":1024.0") || json.contains("\"result\":1024"));
    }

    @Test(expected = Exception.class)
    public void testCalculatorDivisionByZero() throws Exception {
        calculator.execute("10 / 0".getBytes());
    }

    // ===== Timestamp Tests =====

    @Test
    public void testTimestampNow() throws Exception {
        byte[] result = timestamp.execute("now".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"timestamp\""));
        assertTrue(json.contains("\"formatted\""));
    }

    @Test
    public void testTimestampUtc() throws Exception {
        byte[] result = timestamp.execute("utc".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"utc\""));
    }

    @Test
    public void testTimestampFormat() throws Exception {
        byte[] result = timestamp.execute("format".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"formatted\""));
    }

    // ===== JSON Format Tests =====

    @Test
    public void testJsonFormatValid() throws Exception {
        byte[] result = jsonFormat.execute("{\"name\":\"test\",\"value\":123}".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"valid\":true"));
    }

    @Test
    public void testJsonFormatInvalid() throws Exception {
        byte[] result = jsonFormat.execute("{invalid json}".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"valid\":false"));
    }

    @Test
    public void testJsonFormatArray() throws Exception {
        byte[] result = jsonFormat.execute("[1,2,3]".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"type\":\"array\""));
    }

    // ===== Hash Tests =====

    @Test
    public void testHashDefault() throws Exception {
        byte[] result = hash.execute("hello".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"hash\""));
        assertTrue(json.contains("\"algorithm\":\"SHA-256\""));
    }

    @Test
    public void testHashMd5() throws Exception {
        byte[] result = hash.execute("md5:hello".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"algorithm\":\"MD5\""));
        // MD5 of "hello" is known
        assertTrue(json.contains("5d41402abc4b2a76b9719d911017c592"));
    }

    @Test
    public void testHashSha256() throws Exception {
        byte[] result = hash.execute("sha256:hello".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"algorithm\":\"SHA-256\""));
        // SHA-256 produces 64 hex characters
        assertTrue(json.contains("\"hashLength\":64"));
    }

    // ===== Echo Tests =====

    @Test
    public void testEchoBasic() throws Exception {
        byte[] result = echo.execute("test message".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"echo\":\"test message\""));
    }

    @Test
    public void testEchoLength() throws Exception {
        byte[] result = echo.execute("hello".getBytes());
        String json = new String(result);
        assertTrue(json.contains("\"length\":5"));
    }

    // ===== Skill Metadata Tests =====

    @Test
    public void testSkillMetadata() {
        assertEquals("calculator", calculator.getSkillId());
        assertEquals("Calculator", calculator.getSkillName());
        assertEquals("math", calculator.getCategory());

        assertEquals("timestamp", timestamp.getSkillId());
        assertEquals("Timestamp", timestamp.getSkillName());
        assertEquals("time", timestamp.getCategory());

        assertEquals("json.format", jsonFormat.getSkillId());
        assertEquals("JSON Formatter", jsonFormat.getSkillName());
        assertEquals("data", jsonFormat.getCategory());

        assertEquals("hash", hash.getSkillId());
        assertEquals("Hash", hash.getSkillName());
        assertEquals("crypto", hash.getCategory());
    }
}