package com.ofa.agent.automation.webview;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationResult;

import org.json.JSONObject;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * WebFormFiller - automatic form filling for web pages.
 * Supports various input types including text, select, checkbox, radio, and file inputs.
 */
public class WebFormFiller {

    private static final String TAG = "WebFormFiller";

    private final WebViewAutomation webViewAutomation;
    private final JsExecutor jsExecutor;

    // Configuration
    private boolean triggerEvents = true; // Trigger input/change events after filling
    private long fillDelay = 50; // Delay between field fills (ms)

    public WebFormFiller(@NonNull WebViewAutomation webViewAutomation) {
        this.webViewAutomation = webViewAutomation;
        this.jsExecutor = webViewAutomation.getJsExecutor();
    }

    /**
     * Form field definition
     */
    public static class FormField {
        public final String selector;
        public final String value;
        public final String type; // text, select, checkbox, radio, file

        public FormField(String selector, String value, String type) {
            this.selector = selector;
            this.value = value;
            this.type = type;
        }

        public FormField(String selector, String value) {
            this(selector, value, "text");
        }
    }

    /**
     * Form info
     */
    public static class FormInfo {
        public final String id;
        public final String name;
        public final String action;
        public final String method;
        public final int fieldCount;

        public FormInfo(String id, String name, String action, String method, int fieldCount) {
            this.id = id;
            this.name = name;
            this.action = action;
            this.method = method;
            this.fieldCount = fieldCount;
        }
    }

    /**
     * Fill a single text input
     */
    @NonNull
    public AutomationResult fillText(@NonNull String selector, @NonNull String value) {
        String script = buildFillTextScript(selector, value);
        return webViewAutomation.executeJs(script);
    }

    /**
     * Build fill text script
     */
    @NonNull
    private String buildFillTextScript(@NonNull String selector, @NonNull String value) {
        String escapedValue = escapeJsString(value);
        String escapedSelector = escapeJsString(selector);

        if (triggerEvents) {
            return String.format(
                "(function() {" +
                "  var el = document.querySelector('%s');" +
                "  if (el) {" +
                "    el.focus();" +
                "    el.value = '%s';" +
                "    el.dispatchEvent(new Event('input', { bubbles: true }));" +
                "    el.dispatchEvent(new Event('change', { bubbles: true }));" +
                "    el.blur();" +
                "    return true;" +
                "  }" +
                "  return false;" +
                "})()",
                escapedSelector, escapedValue
            );
        } else {
            return String.format(
                "(function() {" +
                "  var el = document.querySelector('%s');" +
                "  if (el) { el.value = '%s'; return true; }" +
                "  return false;" +
                "})()",
                escapedSelector, escapedValue
            );
        }
    }

    /**
     * Fill a select/dropdown
     */
    @NonNull
    public AutomationResult fillSelect(@NonNull String selector, @NonNull String value) {
        String escapedValue = escapeJsString(value);
        String escapedSelector = escapeJsString(selector);

        String script = String.format(
            "(function() {" +
            "  var select = document.querySelector('%s');" +
            "  if (!select) return false;" +
            "  var options = select.options;" +
            "  for (var i = 0; i < options.length; i++) {" +
            "    if (options[i].value === '%s' || options[i].text === '%s') {" +
            "      select.selectedIndex = i;" +
            "      select.dispatchEvent(new Event('change', { bubbles: true }));" +
            "      return true;" +
            "    }" +
            "  }" +
            "  return false;" +
            "})()",
            escapedSelector, escapedValue, escapedValue
        );

        return webViewAutomation.executeJs(script);
    }

    /**
     * Fill a select by index
     */
    @NonNull
    public AutomationResult fillSelectByIndex(@NonNull String selector, int index) {
        String escapedSelector = escapeJsString(selector);

        String script = String.format(
            "(function() {" +
            "  var select = document.querySelector('%s');" +
            "  if (!select || select.options.length <= %d) return false;" +
            "  select.selectedIndex = %d;" +
            "  select.dispatchEvent(new Event('change', { bubbles: true }));" +
            "  return true;" +
            "})()",
            escapedSelector, index, index
        );

        return webViewAutomation.executeJs(script);
    }

    /**
     * Set checkbox state
     */
    @NonNull
    public AutomationResult setCheckbox(@NonNull String selector, boolean checked) {
        String escapedSelector = escapeJsString(selector);

        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('%s');" +
            "  if (el && el.type === 'checkbox') {" +
            "    el.checked = %s;" +
            "    el.dispatchEvent(new Event('change', { bubbles: true }));" +
            "    return true;" +
            "  }" +
            "  return false;" +
            "})()",
            escapedSelector, checked ? "true" : "false"
        );

        return webViewAutomation.executeJs(script);
    }

    /**
     * Set radio button by value
     */
    @NonNull
    public AutomationResult setRadio(@NonNull String name, @NonNull String value) {
        String escapedName = escapeJsString(name);
        String escapedValue = escapeJsString(value);

        String script = String.format(
            "(function() {" +
            "  var radios = document.querySelectorAll('input[type=radio][name=\"%s\"]');" +
            "  for (var i = 0; i < radios.length; i++) {" +
            "    if (radios[i].value === '%s') {" +
            "      radios[i].checked = true;" +
            "      radios[i].dispatchEvent(new Event('change', { bubbles: true }));" +
            "      return true;" +
            "    }" +
            "  }" +
            "  return false;" +
            "})()",
            escapedName, escapedValue
        );

        return webViewAutomation.executeJs(script);
    }

    /**
     * Fill multiple form fields
     */
    @NonNull
    public AutomationResult fillForm(@NonNull Map<String, String> fields) {
        int successCount = 0;
        int failCount = 0;

        for (Map.Entry<String, String> entry : fields.entrySet()) {
            String selector = entry.getKey();
            String value = entry.getValue();

            // Determine field type and fill accordingly
            AutomationResult result = fillFieldAuto(selector, value);

            if (result.isSuccess()) {
                successCount++;
            } else {
                failCount++;
                Log.w(TAG, "Failed to fill field: " + selector);
            }

            // Small delay between fields
            if (fillDelay > 0) {
                try {
                    Thread.sleep(fillDelay);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    break;
                }
            }
        }

        if (failCount == 0) {
            return new AutomationResult("fillForm", 0);
        } else {
            return new AutomationResult("fillForm",
                "Filled " + successCount + "/" + (successCount + failCount) + " fields");
        }
    }

    /**
     * Fill form fields from list
     */
    @NonNull
    public AutomationResult fillFormList(@NonNull List<FormField> fields) {
        int successCount = 0;
        int failCount = 0;

        for (FormField field : fields) {
            AutomationResult result;

            switch (field.type) {
                case "select":
                    result = fillSelect(field.selector, field.value);
                    break;
                case "checkbox":
                    result = setCheckbox(field.selector, Boolean.parseBoolean(field.value));
                    break;
                case "radio":
                    result = setRadio(field.selector, field.value);
                    break;
                default:
                    result = fillFieldAuto(field.selector, field.value);
            }

            if (result.isSuccess()) {
                successCount++;
            } else {
                failCount++;
            }

            if (fillDelay > 0) {
                try {
                    Thread.sleep(fillDelay);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    break;
                }
            }
        }

        if (failCount == 0) {
            return new AutomationResult("fillFormList", 0);
        } else {
            return new AutomationResult("fillFormList",
                "Filled " + successCount + "/" + (successCount + failCount) + " fields");
        }
    }

    /**
     * Auto-detect field type and fill
     */
    @NonNull
    private AutomationResult fillFieldAuto(@NonNull String selector, @NonNull String value) {
        String escapedSelector = escapeJsString(selector);
        String escapedValue = escapeJsString(value);

        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('%s');" +
            "  if (!el) return false;" +
            "  var type = el.type || el.tagName.toLowerCase();" +
            "  switch(type) {" +
            "    case 'select-one':" +
            "    case 'select-multiple':" +
            "      var options = el.options;" +
            "      for (var i = 0; i < options.length; i++) {" +
            "        if (options[i].value === '%s' || options[i].text === '%s') {" +
            "          el.selectedIndex = i;" +
            "          el.dispatchEvent(new Event('change', { bubbles: true }));" +
            "          return true;" +
            "        }" +
            "      }" +
            "      return false;" +
            "    case 'checkbox':" +
            "      el.checked = %s;" +
            "      el.dispatchEvent(new Event('change', { bubbles: true }));" +
            "      return true;" +
            "    default:" +
            "      el.focus();" +
            "      el.value = '%s';" +
            "      el.dispatchEvent(new Event('input', { bubbles: true }));" +
            "      el.dispatchEvent(new Event('change', { bubbles: true }));" +
            "      el.blur();" +
            "      return true;" +
            "  }" +
            "})()",
            escapedSelector, escapedValue, escapedValue,
            Boolean.parseBoolean(value) ? "true" : "false",
            escapedValue
        );

        return webViewAutomation.executeJs(script);
    }

    /**
     * Clear form field
     */
    @NonNull
    public AutomationResult clearField(@NonNull String selector) {
        String escapedSelector = escapeJsString(selector);

        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('%s');" +
            "  if (el) {" +
            "    el.value = '';" +
            "    el.dispatchEvent(new Event('input', { bubbles: true }));" +
            "    return true;" +
            "  }" +
            "  return false;" +
            "})()",
            escapedSelector
        );

        return webViewAutomation.executeJs(script);
    }

    /**
     * Clear all fields in a form
     */
    @NonNull
    public AutomationResult clearForm(@NonNull String formSelector) {
        String escapedSelector = escapeJsString(formSelector);

        String script = String.format(
            "(function() {" +
            "  var form = document.querySelector('%s');" +
            "  if (!form) return false;" +
            "  var inputs = form.querySelectorAll('input, textarea, select');" +
            "  for (var i = 0; i < inputs.length; i++) {" +
            "    var el = inputs[i];" +
            "    if (el.type === 'checkbox' || el.type === 'radio') {" +
            "      el.checked = false;" +
            "    } else if (el.tagName === 'SELECT') {" +
            "      el.selectedIndex = -1;" +
            "    } else {" +
            "      el.value = '';" +
            "    }" +
            "  }" +
            "  return true;" +
            "})()",
            escapedSelector
        );

        return webViewAutomation.executeJs(script);
    }

    /**
     * Submit form
     */
    @NonNull
    public AutomationResult submitForm(@NonNull String formSelector) {
        String escapedSelector = escapeJsString(formSelector);

        String script = String.format(
            "(function() {" +
            "  var form = document.querySelector('%s');" +
            "  if (form && form.tagName === 'FORM') {" +
            "    form.submit();" +
            "    return true;" +
            "  }" +
            "  return false;" +
            "})()",
            escapedSelector
        );

        return webViewAutomation.executeJs(script);
    }

    /**
     * Submit form by clicking submit button
     */
    @NonNull
    public AutomationResult submitFormByButton(@NonNull String buttonSelector) {
        return webViewAutomation.click(buttonSelector);
    }

    /**
     * Get form data as JSON
     */
    @Nullable
    public JSONObject getFormData(@NonNull String formSelector) {
        String escapedSelector = escapeJsString(formSelector);

        String script = String.format(
            "(function() {" +
            "  var form = document.querySelector('%s');" +
            "  if (!form) return null;" +
            "  var data = {};" +
            "  var inputs = form.querySelectorAll('input, textarea, select');" +
            "  for (var i = 0; i < inputs.length; i++) {" +
            "    var el = inputs[i];" +
            "    var name = el.name || el.id;" +
            "    if (!name) continue;" +
            "    if (el.type === 'checkbox') {" +
            "      data[name] = el.checked;" +
            "    } else if (el.type === 'radio') {" +
            "      if (el.checked) data[name] = el.value;" +
            "    } else {" +
            "      data[name] = el.value;" +
            "    }" +
            "  }" +
            "  return JSON.stringify(data);" +
            "})()",
            escapedSelector
        );

        return jsExecutor.executeJson(script);
    }

    /**
     * Get all forms on page
     */
    @Nullable
    public java.util.List<FormInfo> getAllForms() {
        org.json.JSONArray forms = jsExecutor.executeJsonArray(
            "JSON.stringify(Array.from(document.querySelectorAll('form')).map(f => ({" +
            "  id: f.id, name: f.name, action: f.action, method: f.method, " +
            "  fieldCount: f.querySelectorAll('input, textarea, select').length" +
            "})))"
        );

        if (forms == null) return null;

        java.util.List<FormInfo> result = new java.util.ArrayList<>();
        for (int i = 0; i < forms.length(); i++) {
            try {
                JSONObject f = forms.getJSONObject(i);
                result.add(new FormInfo(
                    f.optString("id"),
                    f.optString("name"),
                    f.optString("action"),
                    f.optString("method"),
                    f.optInt("fieldCount")
                ));
            } catch (Exception e) {
                Log.w(TAG, "Error parsing form info", e);
            }
        }

        return result;
    }

    /**
     * Fill form by field name (more resilient than selector)
     */
    @NonNull
    public AutomationResult fillByName(@NonNull String name, @NonNull String value) {
        String escapedName = escapeJsString(name);
        String escapedValue = escapeJsString(value);

        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('[name=\"%s\"]');" +
            "  if (!el) return false;" +
            "  el.focus();" +
            "  el.value = '%s';" +
            "  el.dispatchEvent(new Event('input', { bubbles: true }));" +
            "  el.dispatchEvent(new Event('change', { bubbles: true }));" +
            "  el.blur();" +
            "  return true;" +
            "})()",
            escapedName, escapedValue
        );

        return webViewAutomation.executeJs(script);
    }

    /**
     * Fill form by label text
     */
    @NonNull
    public AutomationResult fillByLabel(@NonNull String labelText, @NonNull String value) {
        String escapedLabel = escapeJsString(labelText);
        String escapedValue = escapeJsString(value);

        String script = String.format(
            "(function() {" +
            "  var labels = document.querySelectorAll('label');" +
            "  for (var i = 0; i < labels.length; i++) {" +
            "    if (labels[i].textContent.trim() === '%s') {" +
            "      var forId = labels[i].getAttribute('for');" +
            "      var el;" +
            "      if (forId) {" +
            "        el = document.getElementById(forId);" +
            "      } else {" +
            "        el = labels[i].querySelector('input, textarea, select');" +
            "      }" +
            "      if (el) {" +
            "        el.focus();" +
            "        el.value = '%s';" +
            "        el.dispatchEvent(new Event('input', { bubbles: true }));" +
            "        el.dispatchEvent(new Event('change', { bubbles: true }));" +
            "        el.blur();" +
            "        return true;" +
            "      }" +
            "    }" +
            "  }" +
            "  return false;" +
            "})()",
            escapedLabel, escapedValue
        );

        return webViewAutomation.executeJs(script);
    }

    /**
     * Escape JavaScript string
     */
    @NonNull
    private String escapeJsString(@NonNull String s) {
        return s.replace("\\", "\\\\")
                .replace("'", "\\'")
                .replace("\"", "\\\"")
                .replace("\n", "\\n")
                .replace("\r", "\\r")
                .replace("\t", "\\t");
    }

    // ===== Configuration =====

    public void setTriggerEvents(boolean trigger) {
        this.triggerEvents = trigger;
    }

    public void setFillDelay(long delayMs) {
        this.fillDelay = delayMs;
    }
}