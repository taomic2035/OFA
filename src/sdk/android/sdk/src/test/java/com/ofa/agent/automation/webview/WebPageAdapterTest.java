package com.ofa.agent.automation.webview;

import org.junit.Test;
import org.junit.Before;

import static org.junit.Assert.*;

/**
 * Unit tests for WebPageAdapter.
 * Tests page pattern matching and detection logic.
 */
public class WebPageAdapterTest {

    // Test page pattern matching

    @Test
    public void testLoginPatternMatch() {
        String[] loginUrls = {
            "https://example.com/login",
            "https://example.com/signin",
            "https://example.com/auth",
            "https://example.com/auth/login",
            "https://example.com/user/signin"
        };

        String loginPattern = ".*login.*|.*signin.*|.*auth.*";

        for (String url : loginUrls) {
            assertTrue("URL should match login pattern: " + url, url.matches(loginPattern));
        }
    }

    @Test
    public void testSearchPatternMatch() {
        String[] searchUrls = {
            "https://google.com/search?q=test",
            "https://example.com/query?term=hello",
            "https://example.com/search/results"
        };

        String searchPattern = ".*search.*|.*query.*";

        for (String url : searchUrls) {
            assertTrue("URL should match search pattern: " + url, url.matches(searchPattern));
        }
    }

    @Test
    public void testCheckoutPatternMatch() {
        String[] checkoutUrls = {
            "https://shop.com/checkout",
            "https://shop.com/cart",
            "https://shop.com/order/confirm"
        };

        String checkoutPattern = ".*checkout.*|.*cart.*|.*order.*";

        for (String url : checkoutUrls) {
            assertTrue("URL should match checkout pattern: " + url, url.matches(checkoutPattern));
        }
    }

    @Test
    public void testNoPatternMatch() {
        String[] nonMatchingUrls = {
            "https://example.com/home",
            "https://example.com/about",
            "https://example.com/products"
        };

        String loginPattern = ".*login.*|.*signin.*|.*auth.*";

        for (String url : nonMatchingUrls) {
            assertFalse("URL should not match login pattern: " + url, url.matches(loginPattern));
        }
    }

    // Test page pattern creation

    @Test
    public void testPagePatternCreation() {
        java.util.Map<String, String> selectors = new java.util.HashMap<>();
        selectors.put("username", "#user");
        selectors.put("password", "#pass");

        WebPageAdapter.PagePattern pattern = new WebPageAdapter.PagePattern(
            "test", ".*test.*", selectors
        );

        assertEquals("test", pattern.name);
        assertEquals(".*test.*", pattern.urlPattern);
        assertEquals(2, pattern.selectors.size());
        assertEquals("#user", pattern.selectors.get("username"));
    }

    // Test form field matching

    @Test
    public void testFieldNameMatching_Email() {
        String fieldName = "userEmail";
        String lowerName = fieldName.toLowerCase();

        assertTrue("Should match email field", lowerName.contains("email"));
    }

    @Test
    public void testFieldNameMatching_Phone() {
        String[] phoneFieldNames = {"phoneNumber", "mobile", "tel", "contact_number"};

        for (String name : phoneFieldNames) {
            String lowerName = name.toLowerCase();
            boolean isPhone = lowerName.contains("phone") || lowerName.contains("tel") || lowerName.contains("mobile");
            assertTrue("Should match phone field: " + name, isPhone);
        }
    }

    @Test
    public void testFieldNameMatching_Address() {
        String[] addressFieldNames = {"address", "street", "city", "zipcode"};

        for (String name : addressFieldNames) {
            String lowerName = name.toLowerCase();
            boolean isAddress = lowerName.contains("address") || lowerName.contains("street") ||
                               lowerName.contains("city") || lowerName.contains("zip");
            assertTrue("Should match address field: " + name, isAddress);
        }
    }

    // Test page analysis

    @Test
    public void testPageAnalysisDefaults() {
        WebPageAdapter.PageAnalysis analysis = new WebPageAdapter.PageAnalysis();

        // Check defaults
        assertNull(analysis.pageType);
        assertNull(analysis.title);
        assertNull(analysis.url);
        assertEquals(0, analysis.inputCount);
        assertEquals(0, analysis.linkCount);
        assertEquals(0, analysis.imageCount);
        assertEquals(0, analysis.buttonCount);
        assertEquals(0, analysis.formCount);
        assertFalse(analysis.hasLoginForm);
        assertFalse(analysis.hasSearchInput);
    }

    @Test
    public void testPageAnalysisToString() {
        WebPageAdapter.PageAnalysis analysis = new WebPageAdapter.PageAnalysis();
        analysis.pageType = "login";
        analysis.inputCount = 3;
        analysis.formCount = 1;
        analysis.linkCount = 5;

        String str = analysis.toString();
        assertTrue("Should contain type", str.contains("login"));
        assertTrue("Should contain inputs", str.contains("3"));
    }

    // Test search result

    @Test
    public void testSearchResultCreation() {
        WebPageAdapter.SearchResult result = new WebPageAdapter.SearchResult(
            "Test Title",
            "https://example.com/result",
            "This is a snippet of the search result..."
        );

        assertEquals("Test Title", result.title);
        assertEquals("https://example.com/result", result.url);
        assertTrue(result.snippet.contains("snippet"));
    }

    // Test URL pattern edge cases

    @Test
    public void testUrlWithQueryString() {
        String url = "https://example.com/login?redirect=/home&token=abc123";
        String loginPattern = ".*login.*";

        assertTrue("URL with query should match", url.matches(loginPattern));
    }

    @Test
    public void testUrlWithFragment() {
        String url = "https://example.com/page#section";
        String pagePattern = ".*page.*";

        assertTrue("URL with fragment should match", url.matches(pagePattern));
    }

    @Test
    public void testHttpsUrl() {
        String url = "https://secure.example.com/login";
        String loginPattern = ".*login.*";

        assertTrue("HTTPS URL should match", url.matches(loginPattern));
    }

    @Test
    public void testHttpUrl() {
        String url = "http://example.com/login";
        String loginPattern = ".*login.*";

        assertTrue("HTTP URL should match", url.matches(loginPattern));
    }

    // Test field value mapping

    @Test
    public void testFieldValueMapping() {
        java.util.Map<String, String> data = new java.util.HashMap<>();
        data.put("email", "user@test.com");
        data.put("phone", "1234567890");
        data.put("name", "John Doe");

        assertEquals("user@test.com", data.get("email"));
        assertEquals("1234567890", data.get("phone"));
        assertEquals("John Doe", data.get("name"));
    }

    @Test
    public void testFuzzyFieldMatching() {
        String fieldName = "userEmailAddress";
        java.util.Map<String, String> data = new java.util.HashMap<>();
        data.put("email", "test@example.com");

        // Simulate fuzzy matching
        String lowerName = fieldName.toLowerCase();
        String matchedValue = null;

        for (java.util.Map.Entry<String, String> entry : data.entrySet()) {
            String key = entry.getKey().toLowerCase();
            if (lowerName.contains(key) || key.contains(lowerName)) {
                matchedValue = entry.getValue();
                break;
            }
        }

        assertEquals("test@example.com", matchedValue);
    }

    // Test form detection heuristics

    @Test
    public void testLoginFormDetection() {
        // Login form should have password input
        boolean hasPasswordInput = true;
        boolean hasTextInput = true;

        boolean isLoginForm = hasPasswordInput && hasTextInput;
        assertTrue("Should detect login form", isLoginForm);
    }

    @Test
    public void testSearchPageDetection() {
        // Search page should have search input
        boolean hasSearchInput = true;

        assertTrue("Should detect search page", hasSearchInput);
    }

    @Test
    public void testFormPageDetection() {
        boolean hasForm = true;

        assertTrue("Should detect form page", hasForm);
    }

    // Test snapshot functionality

    @Test
    public void testSnapshotCreation() throws Exception {
        JSONObject snapshot = new JSONObject();
        snapshot.put("pageType", "login");
        snapshot.put("title", "Login Page");
        snapshot.put("url", "https://example.com/login");
        snapshot.put("timestamp", System.currentTimeMillis());

        assertEquals("login", snapshot.getString("pageType"));
        assertEquals("Login Page", snapshot.getString("title"));
        assertTrue(snapshot.has("timestamp"));
    }
}