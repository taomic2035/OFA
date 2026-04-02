package com.ofa.agent.social.adapter;

import android.content.Context;
import android.database.Cursor;
import android.net.Uri;
import android.provider.ContactsContract;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Contact Adapter - provides access to device contacts.
 *
 * Integrates with Android ContactsContract to:
 * - Find contacts by name
 * - Get contact phone numbers
 * - Search contacts
 * - Get contact details (name, phone, email, etc.)
 */
public class ContactAdapter {

    private static final String TAG = "ContactAdapter";

    private final Context context;

    /**
     * Contact information
     */
    public static class ContactInfo {
        public final String id;
        public final String displayName;
        public final List<String> phoneNumbers;
        public final List<String> emails;
        public final String photoUri;
        public final String company;
        public final String notes;

        // Social handles (from contact notes or custom fields)
        public final Map<String, String> socialHandles;

        public ContactInfo(String id, String displayName, List<String> phoneNumbers,
                           List<String> emails, String photoUri, String company,
                           String notes, Map<String, String> socialHandles) {
            this.id = id;
            this.displayName = displayName;
            this.phoneNumbers = phoneNumbers;
            this.emails = emails;
            this.photoUri = photoUri;
            this.company = company;
            this.notes = notes;
            this.socialHandles = socialHandles;
        }

        /**
         * Get primary phone number
         */
        @Nullable
        public String getPrimaryPhone() {
            if (phoneNumbers.isEmpty()) return null;
            // Prefer mobile number
            for (String phone : phoneNumbers) {
                if (phone.startsWith("1") && phone.length() == 11) {
                    return phone;
                }
            }
            return phoneNumbers.get(0);
        }

        /**
         * Get primary email
         */
        @Nullable
        public String getPrimaryEmail() {
            return emails.isEmpty() ? null : emails.get(0);
        }

        /**
         * Get WeChat ID if stored
         */
        @Nullable
        public String getWeChatId() {
            return socialHandles.get("wechat");
        }

        /**
         * Check if has phone number
         */
        public boolean hasPhone() {
            return !phoneNumbers.isEmpty();
        }

        /**
         * Check if has email
         */
        public boolean hasEmail() {
            return !emails.isEmpty();
        }

        @NonNull
        @Override
        public String toString() {
            return String.format("Contact{name='%s', phones=%s, wechat='%s'}",
                displayName, phoneNumbers, socialHandles.get("wechat"));
        }
    }

    /**
     * Create contact adapter
     */
    public ContactAdapter(@NonNull Context context) {
        this.context = context;
    }

    /**
     * Find contact by name (exact or partial match)
     */
    @Nullable
    public ContactInfo findContact(@NonNull String name) {
        // Try exact match first
        ContactInfo exact = findContactExact(name);
        if (exact != null) {
            return exact;
        }

        // Try partial match
        List<ContactInfo> partials = searchContacts(name);
        if (!partials.isEmpty()) {
            return partials.get(0);
        }

        return null;
    }

    /**
     * Find contact by exact name
     */
    @Nullable
    private ContactInfo findContactExact(@NonNull String name) {
        Cursor cursor = null;
        try {
            Uri uri = ContactsContract.Contacts.CONTENT_URI;
            String selection = ContactsContract.Contacts.DISPLAY_NAME + " = ?";
            String[] args = {name};

            cursor = context.getContentResolver().query(uri, null, selection, args, null);

            if (cursor != null && cursor.moveToFirst()) {
                return extractContact(cursor);
            }
        } catch (Exception e) {
            Log.w(TAG, "Error finding contact: " + e.getMessage());
        } finally {
            if (cursor != null) cursor.close();
        }

        return null;
    }

    /**
     * Search contacts by query
     */
    @NonNull
    public List<ContactInfo> searchContacts(@NonNull String query) {
        List<ContactInfo> results = new ArrayList<>();
        Cursor cursor = null;

        try {
            Uri uri = ContactsContract.Contacts.CONTENT_URI;
            String selection = ContactsContract.Contacts.DISPLAY_NAME + " LIKE ?";
            String[] args = {"%" + query + "%"};

            cursor = context.getContentResolver().query(uri, null, selection, args,
                ContactsContract.Contacts.DISPLAY_NAME + " ASC LIMIT 20");

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    ContactInfo contact = extractContact(cursor);
                    if (contact != null) {
                        results.add(contact);
                    }
                }
            }
        } catch (Exception e) {
            Log.w(TAG, "Error searching contacts: " + e.getMessage());
        } finally {
            if (cursor != null) cursor.close();
        }

        return results;
    }

    /**
     * Get all contacts (paginated)
     */
    @NonNull
    public List<ContactInfo> getAllContacts() {
        return getContacts(0, 100);
    }

    /**
     * Get contacts with pagination
     */
    @NonNull
    public List<ContactInfo> getContacts(int offset, int limit) {
        List<ContactInfo> results = new ArrayList<>();
        Cursor cursor = null;

        try {
            Uri uri = ContactsContract.Contacts.CONTENT_URI;
            String sortOrder = ContactsContract.Contacts.DISPLAY_NAME + " ASC LIMIT " + limit +
                " OFFSET " + offset;

            cursor = context.getContentResolver().query(uri, null, null, null, sortOrder);

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    ContactInfo contact = extractContact(cursor);
                    if (contact != null) {
                        results.add(contact);
                    }
                }
            }
        } catch (Exception e) {
            Log.w(TAG, "Error getting contacts: " + e.getMessage());
        } finally {
            if (cursor != null) cursor.close();
        }

        return results;
    }

    /**
     * Get contact by ID
     */
    @Nullable
    public ContactInfo getContactById(@NonNull String contactId) {
        Cursor cursor = null;

        try {
            Uri uri = ContactsContract.Contacts.CONTENT_URI;
            String selection = ContactsContract.Contacts._ID + " = ?";
            String[] args = {contactId};

            cursor = context.getContentResolver().query(uri, null, selection, args, null);

            if (cursor != null && cursor.moveToFirst()) {
                return extractContact(cursor);
            }
        } catch (Exception e) {
            Log.w(TAG, "Error getting contact by ID: " + e.getMessage());
        } finally {
            if (cursor != null) cursor.close();
        }

        return null;
    }

    /**
     * Extract contact info from cursor
     */
    @Nullable
    private ContactInfo extractContact(@NonNull Cursor cursor) {
        try {
            String id = cursor.getString(cursor.getColumnIndex(ContactsContract.Contacts._ID));
            String displayName = cursor.getString(cursor.getColumnIndex(
                ContactsContract.Contacts.DISPLAY_NAME));
            String photoUri = cursor.getString(cursor.getColumnIndex(
                ContactsContract.Contacts.PHOTO_URI));

            // Get phone numbers
            List<String> phoneNumbers = getPhoneNumbers(id);

            // Get emails
            List<String> emails = getEmails(id);

            // Get organization/company
            String company = getCompany(id);

            // Get notes
            String notes = getNotes(id);

            // Parse social handles from notes or structured fields
            Map<String, String> socialHandles = parseSocialHandles(notes);

            return new ContactInfo(id, displayName, phoneNumbers, emails,
                photoUri, company, notes, socialHandles);
        } catch (Exception e) {
            Log.w(TAG, "Error extracting contact: " + e.getMessage());
            return null;
        }
    }

    /**
     * Get phone numbers for contact
     */
    @NonNull
    private List<String> getPhoneNumbers(@NonNull String contactId) {
        List<String> numbers = new ArrayList<>();
        Cursor cursor = null;

        try {
            Uri uri = ContactsContract.CommonDataKinds.Phone.CONTENT_URI;
            String selection = ContactsContract.CommonDataKinds.Phone.CONTACT_ID + " = ?";
            String[] args = {contactId};

            cursor = context.getContentResolver().query(uri, null, selection, args, null);

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    String number = cursor.getString(cursor.getColumnIndex(
                        ContactsContract.CommonDataKinds.Phone.NUMBER));
                    // Normalize number
                    number = normalizePhoneNumber(number);
                    if (number != null && !numbers.contains(number)) {
                        numbers.add(number);
                    }
                }
            }
        } catch (Exception e) {
            Log.w(TAG, "Error getting phone numbers: " + e.getMessage());
        } finally {
            if (cursor != null) cursor.close();
        }

        return numbers;
    }

    /**
     * Get emails for contact
     */
    @NonNull
    private List<String> getEmails(@NonNull String contactId) {
        List<String> emails = new ArrayList<>();
        Cursor cursor = null;

        try {
            Uri uri = ContactsContract.CommonDataKinds.Email.CONTENT_URI;
            String selection = ContactsContract.CommonDataKinds.Email.CONTACT_ID + " = ?";
            String[] args = {contactId};

            cursor = context.getContentResolver().query(uri, null, selection, args, null);

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    String email = cursor.getString(cursor.getColumnIndex(
                        ContactsContract.CommonDataKinds.Email.ADDRESS));
                    if (email != null && !emails.contains(email)) {
                        emails.add(email);
                    }
                }
            }
        } catch (Exception e) {
            Log.w(TAG, "Error getting emails: " + e.getMessage());
        } finally {
            if (cursor != null) cursor.close();
        }

        return emails;
    }

    /**
     * Get company/organization
     */
    @Nullable
    private String getCompany(@NonNull String contactId) {
        Cursor cursor = null;

        try {
            Uri uri = ContactsContract.Data.CONTENT_URI;
            String selection = ContactsContract.Data.CONTACT_ID + " = ? AND " +
                ContactsContract.Data.MIMETYPE + " = ?";
            String[] args = {contactId, ContactsContract.CommonDataKinds.Organization.CONTENT_ITEM_TYPE};

            cursor = context.getContentResolver().query(uri, null, selection, args, null);

            if (cursor != null && cursor.moveToFirst()) {
                return cursor.getString(cursor.getColumnIndex(
                    ContactsContract.CommonDataKinds.Organization.COMPANY));
            }
        } catch (Exception e) {
            Log.w(TAG, "Error getting company: " + e.getMessage());
        } finally {
            if (cursor != null) cursor.close();
        }

        return null;
    }

    /**
     * Get contact notes
     */
    @Nullable
    private String getNotes(@NonNull String contactId) {
        Cursor cursor = null;

        try {
            Uri uri = ContactsContract.Data.CONTENT_URI;
            String selection = ContactsContract.Data.CONTACT_ID + " = ? AND " +
                ContactsContract.Data.MIMETYPE + " = ?";
            String[] args = {contactId, ContactsContract.CommonDataKinds.Note.CONTENT_ITEM_TYPE};

            cursor = context.getContentResolver().query(uri, null, selection, args, null);

            if (cursor != null && cursor.moveToFirst()) {
                return cursor.getString(cursor.getColumnIndex(
                    ContactsContract.CommonDataKinds.Note.NOTE));
            }
        } catch (Exception e) {
            Log.w(TAG, "Error getting notes: " + e.getMessage());
        } finally {
            if (cursor != null) cursor.close();
        }

        return null;
    }

    /**
     * Parse social handles from notes
     * Notes format example: "微信: abc123\n抖音: @user\n小红书: @creator"
     */
    @NonNull
    private Map<String, String> parseSocialHandles(@Nullable String notes) {
        Map<String, String> handles = new HashMap<>();

        if (notes == null || notes.isEmpty()) {
            return handles;
        }

        // Parse common social platform patterns
        String[] lines = notes.split("\n");
        for (String line : lines) {
            line = line.trim().toLowerCase();

            if (line.contains("微信") || line.contains("wechat")) {
                String[] parts = line.split(":");
                if (parts.length > 1) {
                    handles.put("wechat", parts[1].trim());
                }
            } else if (line.contains("抖音") || line.contains("douyin")) {
                String[] parts = line.split(":");
                if (parts.length > 1) {
                    handles.put("douyin", parts[1].trim());
                }
            } else if (line.contains("小红书") || line.contains("red") || line.contains("xiaohongshu")) {
                String[] parts = line.split(":");
                if (parts.length > 1) {
                    handles.put("xiaohongshu", parts[1].trim());
                }
            } else if (line.contains("qq")) {
                String[] parts = line.split(":");
                if (parts.length > 1) {
                    handles.put("qq", parts[1].trim());
                }
            } else if (line.contains("钉钉") || line.contains("dingtalk")) {
                String[] parts = line.split(":");
                if (parts.length > 1) {
                    handles.put("dingtalk", parts[1].trim());
                }
            } else if (line.contains("支付宝") || line.contains("alipay")) {
                String[] parts = line.split(":");
                if (parts.length > 1) {
                    handles.put("alipay", parts[1].trim());
                }
            }
        }

        return handles;
    }

    /**
     * Normalize phone number (remove spaces, dashes, etc.)
     */
    @Nullable
    private String normalizePhoneNumber(@Nullable String number) {
        if (number == null) return null;

        // Remove non-digit characters except +
        number = number.replaceAll("[^0-9+]", "");

        // Handle Chinese phone numbers
        if (number.length() > 11 && number.startsWith("+86")) {
            number = number.substring(3);
        }

        return number;
    }

    /**
     * Get contact count
     */
    public int getContactCount() {
        Cursor cursor = null;

        try {
            Uri uri = ContactsContract.Contacts.CONTENT_URI;
            cursor = context.getContentResolver().query(uri, null, null, null, null);

            if (cursor != null) {
                return cursor.getCount();
            }
        } catch (Exception e) {
            Log.w(TAG, "Error counting contacts: " + e.getMessage());
        } finally {
            if (cursor != null) cursor.close();
        }

        return 0;
    }

    /**
     * Find contact by phone number
     */
    @Nullable
    public ContactInfo findContactByPhone(@NonNull String phoneNumber) {
        phoneNumber = normalizePhoneNumber(phoneNumber);
        if (phoneNumber == null) return null;

        Cursor cursor = null;

        try {
            Uri uri = ContactsContract.CommonDataKinds.Phone.CONTENT_URI;
            String selection = ContactsContract.CommonDataKinds.Phone.NUMBER + " LIKE ?";
            String[] args = {"%" + phoneNumber + "%"};

            cursor = context.getContentResolver().query(uri, null, selection, args, null);

            if (cursor != null && cursor.moveToFirst()) {
                String contactId = cursor.getString(cursor.getColumnIndex(
                    ContactsContract.CommonDataKinds.Phone.CONTACT_ID));
                return getContactById(contactId);
            }
        } catch (Exception e) {
            Log.w(TAG, "Error finding contact by phone: " + e.getMessage());
        } finally {
            if (cursor != null) cursor.close();
        }

        return null;
    }
}