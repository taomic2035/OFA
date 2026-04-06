package com.ofa.agent.automation.webview;

import android.content.Context;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.tool.ToolDefinition;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolRegistry;

import org.json.JSONObject;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * WebView Tools - provides tools for WebView automation.
 * Registers tools for web automation with the tool registry.
 */
public class WebViewTools {

    private static final String NAMESPACE = "web";

    /**
     * Register all WebView tools
     */
    public static void registerAll(@NonNull ToolRegistry registry, @NonNull WebView webView) {
        WebViewAutomation automation = new WebViewAutomation(webView);

        registry.register(createExecuteJsTool(automation));
        registry.register(createClickTool(automation));
        registry.register(createFillFormTool(automation));
        registry.register(createGetValueTool(automation));
        registry.register(createWaitForLoadTool(automation));
        registry.register(createGetContentTool(automation));
        registry.register(createScrollTool(automation));
        registry.register(createWaitForElementTool(automation));
    }

    /**
     * Create web.executeJs tool
     */
    @NonNull
    private static ToolExecutor createExecuteJsTool(@NonNull WebViewAutomation automation) {
        return new ToolExecutor() {
            @NonNull
            @Override
            public String getName() {
                return NAMESPACE + ".executeJs";
            }

            @NonNull
            @Override
            public ToolDefinition getDefinition() {
                return ToolDefinition.builder()
                    .name(getName())
                    .description("Execute JavaScript code in WebView")
                    .parameter("script", "string", "JavaScript code to execute", true)
                    .parameter("timeout", "number", "Timeout in milliseconds", false)
                    .build();
            }

            @NonNull
            @Override
            public AutomationResult execute(@NonNull Map<String, String> params) {
                String script = params.get("script");
                if (script == null) {
                    return new AutomationResult(getName(), "Missing script parameter");
                }

                long timeout = Long.parseLong(params.getOrDefault("timeout", "10000"));
                return automation.executeJsForResult(script, timeout);
            }

            @Override
            public long getEstimatedTimeMs() {
                return 500;
            }
        };
    }

    /**
     * Create web.click tool
     */
    @NonNull
    private static ToolExecutor createClickTool(@NonNull WebViewAutomation automation) {
        return new ToolExecutor() {
            @NonNull
            @Override
            public String getName() {
                return NAMESPACE + ".click";
            }

            @NonNull
            @Override
            public ToolDefinition getDefinition() {
                return ToolDefinition.builder()
                    .name(getName())
                    .description("Click element in WebView by CSS selector")
                    .parameter("selector", "string", "CSS selector for element", true)
                    .build();
            }

            @NonNull
            @Override
            public AutomationResult execute(@NonNull Map<String, String> params) {
                String selector = params.get("selector");
                if (selector == null) {
                    return new AutomationResult(getName(), "Missing selector parameter");
                }

                return automation.click(selector);
            }

            @Override
            public long getEstimatedTimeMs() {
                return 200;
            }
        };
    }

    /**
     * Create web.fillForm tool
     */
    @NonNull
    private static ToolExecutor createFillFormTool(@NonNull WebViewAutomation automation) {
        WebFormFiller formFiller = new WebFormFiller(automation);

        return new ToolExecutor() {
            @NonNull
            @Override
            public String getName() {
                return NAMESPACE + ".fillForm";
            }

            @NonNull
            @Override
            public ToolDefinition getDefinition() {
                return ToolDefinition.builder()
                    .name(getName())
                    .description("Fill form fields in WebView")
                    .parameter("fields", "object", "Map of selector -> value pairs", true)
                    .build();
            }

            @NonNull
            @Override
            public AutomationResult execute(@NonNull Map<String, String> params) {
                String fieldsJson = params.get("fields");
                if (fieldsJson == null) {
                    return new AutomationResult(getName(), "Missing fields parameter");
                }

                try {
                    JSONObject fieldsObj = new JSONObject(fieldsJson);
                    Map<String, String> fields = new HashMap<>();
                    for (java.util.Iterator<String> it = fieldsObj.keys(); it.hasNext(); ) {
                        String key = it.next();
                        fields.put(key, fieldsObj.getString(key));
                    }

                    return formFiller.fillForm(fields);
                } catch (Exception e) {
                    return new AutomationResult(getName(), "Invalid fields JSON: " + e.getMessage());
                }
            }

            @Override
            public long getEstimatedTimeMs() {
                return 1000;
            }
        };
    }

    /**
     * Create web.getValue tool
     */
    @NonNull
    private static ToolExecutor createGetValueTool(@NonNull WebViewAutomation automation) {
        return new ToolExecutor() {
            @NonNull
            @Override
            public String getName() {
                return NAMESPACE + ".getValue";
            }

            @NonNull
            @Override
            public ToolDefinition getDefinition() {
                return ToolDefinition.builder()
                    .name(getName())
                    .description("Get value from element in WebView")
                    .parameter("selector", "string", "CSS selector for element", true)
                    .parameter("attribute", "string", "Attribute to get (value, text, href, etc.)", false)
                    .build();
            }

            @NonNull
            @Override
            public AutomationResult execute(@NonNull Map<String, String> params) {
                String selector = params.get("selector");
                if (selector == null) {
                    return new AutomationResult(getName(), "Missing selector parameter");
                }

                String attribute = params.getOrDefault("attribute", "value");
                String result;

                switch (attribute) {
                    case "text":
                        result = automation.getText(selector);
                        break;
                    case "value":
                        result = automation.getValue(selector);
                        break;
                    default:
                        result = automation.getAttribute(selector, attribute);
                }

                if (result != null) {
                    try {
                        JSONObject data = new JSONObject();
                        data.put("value", result);
                        return new AutomationResult(getName(), data, 0);
                    } catch (Exception e) {
                        return new AutomationResult(getName(), 0);
                    }
                }

                return new AutomationResult(getName(), "Element not found or value is null");
            }

            @Override
            public long getEstimatedTimeMs() {
                return 200;
            }
        };
    }

    /**
     * Create web.waitForLoad tool
     */
    @NonNull
    private static ToolExecutor createWaitForLoadTool(@NonNull WebViewAutomation automation) {
        return new ToolExecutor() {
            @NonNull
            @Override
            public String getName() {
                return NAMESPACE + ".waitForLoad";
            }

            @NonNull
            @Override
            public ToolDefinition getDefinition() {
                return ToolDefinition.builder()
                    .name(getName())
                    .description("Wait for WebView page to load")
                    .parameter("timeout", "number", "Timeout in milliseconds", false)
                    .build();
            }

            @NonNull
            @Override
            public AutomationResult execute(@NonNull Map<String, String> params) {
                long timeout = Long.parseLong(params.getOrDefault("timeout", "10000"));
                return automation.waitForPageLoad(timeout);
            }

            @Override
            public long getEstimatedTimeMs() {
                return 3000;
            }
        };
    }

    /**
     * Create web.getContent tool
     */
    @NonNull
    private static ToolExecutor createGetContentTool(@NonNull WebViewAutomation automation) {
        return new ToolExecutor() {
            @NonNull
            @Override
            public String getName() {
                return NAMESPACE + ".getContent";
            }

            @NonNull
            @Override
            public ToolDefinition getDefinition() {
                return ToolDefinition.builder()
                    .name(getName())
                    .description("Get page content including inputs, links, and buttons")
                    .build();
            }

            @NonNull
            @Override
            public AutomationResult execute(@NonNull Map<String, String> params) {
                JSONObject content = automation.getPageContent();
                return new AutomationResult(getName(), content, 0);
            }

            @Override
            public long getEstimatedTimeMs() {
                return 500;
            }
        };
    }

    /**
     * Create web.scroll tool
     */
    @NonNull
    private static ToolExecutor createScrollTool(@NonNull WebViewAutomation automation) {
        return new ToolExecutor() {
            @NonNull
            @Override
            public String getName() {
                return NAMESPACE + ".scroll";
            }

            @NonNull
            @Override
            public ToolDefinition getDefinition() {
                return ToolDefinition.builder()
                    .name(getName())
                    .description("Scroll WebView page")
                    .parameter("direction", "string", "Direction: top, bottom, or coordinates x,y", true)
                    .build();
            }

            @NonNull
            @Override
            public AutomationResult execute(@NonNull Map<String, String> params) {
                String direction = params.get("direction");
                if (direction == null) {
                    return new AutomationResult(getName(), "Missing direction parameter");
                }

                switch (direction.toLowerCase()) {
                    case "top":
                        return automation.scrollToTop();
                    case "bottom":
                        return automation.scrollToBottom();
                    default:
                        // Try to parse as coordinates
                        try {
                            String[] coords = direction.split(",");
                            if (coords.length == 2) {
                                int x = Integer.parseInt(coords[0].trim());
                                int y = Integer.parseInt(coords[1].trim());
                                return automation.scroll(x, y);
                            }
                        } catch (Exception e) {
                            // Ignore
                        }
                        return new AutomationResult(getName(), "Invalid direction: " + direction);
                }
            }

            @Override
            public long getEstimatedTimeMs() {
                return 100;
            }
        };
    }

    /**
     * Create web.waitForElement tool
     */
    @NonNull
    private static ToolExecutor createWaitForElementTool(@NonNull WebViewAutomation automation) {
        return new ToolExecutor() {
            @NonNull
            @Override
            public String getName() {
                return NAMESPACE + ".waitForElement";
            }

            @NonNull
            @Override
            public ToolDefinition getDefinition() {
                return ToolDefinition.builder()
                    .name(getName())
                    .description("Wait for element to appear in WebView")
                    .parameter("selector", "string", "CSS selector for element", true)
                    .parameter("timeout", "number", "Timeout in milliseconds", false)
                    .build();
            }

            @NonNull
            @Override
            public AutomationResult execute(@NonNull Map<String, String> params) {
                String selector = params.get("selector");
                if (selector == null) {
                    return new AutomationResult(getName(), "Missing selector parameter");
                }

                long timeout = Long.parseLong(params.getOrDefault("timeout", "10000"));
                return automation.waitForElement(selector, timeout);
            }

            @Override
            public long getEstimatedTimeMs() {
                return 2000;
            }
        };
    }
}