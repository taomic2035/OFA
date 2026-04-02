package com.ofa.agent.tool.builtin.data;

import android.Manifest;
import android.content.ContentResolver;
import android.content.Context;
import android.database.Cursor;
import android.provider.ContactsContract;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.core.content.ContextCompat;

import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.PermissionManager;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.Map;

/**
 * Contacts Tool - query and manage contacts.
 */
public class ContactsTool implements ToolExecutor {

    private static final String TAG = "ContactsTool";

    private final Context context;
    private final ContentResolver contentResolver;

    public ContactsTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.contentResolver = context.getContentResolver();
    }

    @NonNull
    @Override
    public String getToolId() {
        return "contacts";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Query and manage contacts";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "query");

        switch (operation.toLowerCase()) {
            case "query":
                return executeQuery(args, ctx);
            case "search":
                return executeSearch(args, ctx);
            case "count":
                return executeCount(ctx);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return contentResolver != null;
    }

    @Override
    public boolean requiresAuth() {
        return false;
    }

    @Override
    public boolean supportsOffline() {
        return true;
    }

    @Nullable
    @Override
    public String[] getRequiredPermissions() {
        return PermissionManager.getContactsPermissions();
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        return true;
    }

    @Override
    public int getEstimatedTimeMs() {
        return 500;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getQueryDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'query'");
            operation.put("default", "query");
            props.put("operation", operation);

            JSONObject limit = new JSONObject();
            limit.put("type", "integer");
            limit.put("description", "Maximum number of contacts to return");
            limit.put("default", 50);
            props.put("limit", limit);

            JSONObject offset = new JSONObject();
            offset.put("type", "integer");
            offset.put("description", "Offset for pagination");
            offset.put("default", 0);
            props.put("offset", offset);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, null);
        return new ToolDefinition("contacts.query", "Query contacts list",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getSearchDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'search'");
            props.put("operation", operation);

            JSONObject query = new JSONObject();
            query.put("type", "string");
            query.put("description", "Search query string");
            props.put("query", query);

            JSONObject limit = new JSONObject();
            limit.put("type", "integer");
            limit.put("description", "Maximum results");
            limit.put("default", 20);
            props.put("limit", limit);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"query"});
        return new ToolDefinition("contacts.search", "Search contacts",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getCountDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'count'");
        return new ToolDefinition("contacts.count", "Get contacts count",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeQuery(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        if (!hasContactsPermission()) {
            return new ToolResult(getToolId(), "Contacts permission not granted");
        }

        int limit = getIntArg(args, "limit", 50);
        int offset = getIntArg(args, "offset", 0);

        try {
            JSONArray contactsArray = new JSONArray();

            Cursor cursor = contentResolver.query(
                    ContactsContract.Contacts.CONTENT_URI,
                    null, null, null,
                    ContactsContract.Contacts.DISPLAY_NAME + " ASC LIMIT " + limit + " OFFSET " + offset);

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    JSONObject contact = new JSONObject();

                    String id = cursor.getString(cursor.getColumnIndexOrThrow(ContactsContract.Contacts._ID));
                    String name = cursor.getString(cursor.getColumnIndexOrThrow(ContactsContract.Contacts.DISPLAY_NAME));

                    contact.put("id", id);
                    contact.put("name", name);

                    // Get phone numbers
                    JSONArray phones = getPhoneNumbers(id);
                    if (phones.length() > 0) {
                        contact.put("phones", phones);
                    }

                    // Get email addresses
                    JSONArray emails = getEmailAddresses(id);
                    if (emails.length() > 0) {
                        contact.put("emails", emails);
                    }

                    contactsArray.put(contact);
                }
                cursor.close();
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("count", contactsArray.length());
            output.put("contacts", contactsArray);
            output.put("limit", limit);
            output.put("offset", offset);

            return new ToolResult(getToolId(), output, 300);

        } catch (Exception e) {
            Log.e(TAG, "Query contacts failed", e);
            return new ToolResult(getToolId(), "Query failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeSearch(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        if (!hasContactsPermission()) {
            return new ToolResult(getToolId(), "Contacts permission not granted");
        }

        String query = getStringArg(args, "query", null);
        if (query == null || query.isEmpty()) {
            return new ToolResult(getToolId(), "Missing query parameter");
        }

        int limit = getIntArg(args, "limit", 20);

        try {
            JSONArray contactsArray = new JSONArray();

            String selection = ContactsContract.Contacts.DISPLAY_NAME + " LIKE ?";
            String[] selectionArgs = new String[]{"%" + query + "%"};

            Cursor cursor = contentResolver.query(
                    ContactsContract.Contacts.CONTENT_URI,
                    null, selection, selectionArgs,
                    ContactsContract.Contacts.DISPLAY_NAME + " ASC LIMIT " + limit);

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    JSONObject contact = new JSONObject();

                    String id = cursor.getString(cursor.getColumnIndexOrThrow(ContactsContract.Contacts._ID));
                    String name = cursor.getString(cursor.getColumnIndexOrThrow(ContactsContract.Contacts.DISPLAY_NAME));

                    contact.put("id", id);
                    contact.put("name", name);

                    JSONArray phones = getPhoneNumbers(id);
                    if (phones.length() > 0) {
                        contact.put("phones", phones);
                    }

                    JSONArray emails = getEmailAddresses(id);
                    if (emails.length() > 0) {
                        contact.put("emails", emails);
                    }

                    contactsArray.put(contact);
                }
                cursor.close();
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("query", query);
            output.put("count", contactsArray.length());
            output.put("contacts", contactsArray);

            return new ToolResult(getToolId(), output, 300);

        } catch (Exception e) {
            Log.e(TAG, "Search contacts failed", e);
            return new ToolResult(getToolId(), "Search failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeCount(@NonNull ExecutionContext ctx) {
        if (!hasContactsPermission()) {
            return new ToolResult(getToolId(), "Contacts permission not granted");
        }

        try {
            Cursor cursor = contentResolver.query(
                    ContactsContract.Contacts.CONTENT_URI,
                    null, null, null, null);

            int count = 0;
            if (cursor != null) {
                count = cursor.getCount();
                cursor.close();
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("count", count);

            return new ToolResult(getToolId(), output, 100);

        } catch (Exception e) {
            Log.e(TAG, "Count contacts failed", e);
            return new ToolResult(getToolId(), "Count failed: " + e.getMessage());
        }
    }

    // ===== Helper Methods =====

    private boolean hasContactsPermission() {
        return ContextCompat.checkSelfPermission(context, Manifest.permission.READ_CONTACTS)
                == android.content.pm.PackageManager.PERMISSION_GRANTED;
    }

    @NonNull
    private JSONArray getPhoneNumbers(@NonNull String contactId) {
        JSONArray phones = new JSONArray();

        Cursor phoneCursor = contentResolver.query(
                ContactsContract.CommonDataKinds.Phone.CONTENT_URI,
                null,
                ContactsContract.CommonDataKinds.Phone.CONTACT_ID + " = ?",
                new String[]{contactId},
                null);

        if (phoneCursor != null) {
            while (phoneCursor.moveToNext()) {
                String number = phoneCursor.getString(phoneCursor.getColumnIndexOrThrow(
                        ContactsContract.CommonDataKinds.Phone.NUMBER));
                int type = phoneCursor.getInt(phoneCursor.getColumnIndexOrThrow(
                        ContactsContract.CommonDataKinds.Phone.TYPE));

                try {
                    JSONObject phone = new JSONObject();
                    phone.put("number", number);
                    phone.put("type", getPhoneTypeName(type));
                    phones.put(phone);
                } catch (Exception e) {
                    // Skip on error
                }
            }
            phoneCursor.close();
        }

        return phones;
    }

    @NonNull
    private JSONArray getEmailAddresses(@NonNull String contactId) {
        JSONArray emails = new JSONArray();

        Cursor emailCursor = contentResolver.query(
                ContactsContract.CommonDataKinds.Email.CONTENT_URI,
                null,
                ContactsContract.CommonDataKinds.Email.CONTACT_ID + " = ?",
                new String[]{contactId},
                null);

        if (emailCursor != null) {
            while (emailCursor.moveToNext()) {
                String address = emailCursor.getString(emailCursor.getColumnIndexOrThrow(
                        ContactsContract.CommonDataKinds.Email.ADDRESS));
                int type = emailCursor.getInt(emailCursor.getColumnIndexOrThrow(
                        ContactsContract.CommonDataKinds.Email.TYPE));

                try {
                    JSONObject email = new JSONObject();
                    email.put("address", address);
                    email.put("type", getEmailTypeName(type));
                    emails.put(email);
                } catch (Exception e) {
                    // Skip on error
                }
            }
            emailCursor.close();
        }

        return emails;
    }

    @NonNull
    private String getPhoneTypeName(int type) {
        switch (type) {
            case ContactsContract.CommonDataKinds.Phone.TYPE_HOME: return "home";
            case ContactsContract.CommonDataKinds.Phone.TYPE_MOBILE: return "mobile";
            case ContactsContract.CommonDataKinds.Phone.TYPE_WORK: return "work";
            case ContactsContract.CommonDataKinds.Phone.TYPE_FAX_HOME: return "fax_home";
            case ContactsContract.CommonDataKinds.Phone.TYPE_FAX_WORK: return "fax_work";
            default: return "other";
        }
    }

    @NonNull
    private String getEmailTypeName(int type) {
        switch (type) {
            case ContactsContract.CommonDataKinds.Email.TYPE_HOME: return "home";
            case ContactsContract.CommonDataKinds.Email.TYPE_WORK: return "work";
            case ContactsContract.CommonDataKinds.Email.TYPE_MOBILE: return "mobile";
            case ContactsContract.CommonDataKinds.Email.TYPE_OTHER: return "other";
            default: return "unknown";
        }
    }

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
    }

    private int getIntArg(@NonNull Map<String, Object> args, @NonNull String key, int defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        if (value instanceof Number) return ((Number) value).intValue();
        try {
            return Integer.parseInt(value.toString());
        } catch (NumberFormatException e) {
            return defaultVal;
        }
    }
}