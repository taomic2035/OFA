package com.ofa.agent.security;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

/**
 * 加密数据模型 (v3.7.0)
 */
public class EncryptedData {

    private String dataId;
    private String algorithm;
    private String keyId;
    private String iv;
    private String ciphertext;
    private String tag;
    private long timestamp;
    private String dataType;
    private String checksum;

    public EncryptedData() {
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static EncryptedData fromJson(@NonNull JSONObject json) throws JSONException {
        EncryptedData data = new EncryptedData();
        data.dataId = json.optString("data_id");
        data.algorithm = json.optString("algorithm");
        data.keyId = json.optString("key_id");
        data.iv = json.optString("iv");
        data.ciphertext = json.optString("ciphertext");
        data.tag = json.optString("tag");
        data.timestamp = json.optLong("timestamp", 0);
        data.dataType = json.optString("data_type");
        data.checksum = json.optString("checksum");

        return data;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("data_id", dataId);
            json.put("algorithm", algorithm);
            json.put("key_id", keyId);
            json.put("iv", iv);
            json.put("ciphertext", ciphertext);
            json.put("tag", tag);
            json.put("timestamp", timestamp);
            json.put("data_type", dataType);
            json.put("checksum", checksum);
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    /**
     * 转换为 Map
     */
    @NonNull
    public JSONObject toMap() {
        return toJson();
    }

    /**
     * 验证数据完整性
     */
    public boolean isValid() {
        return dataId != null && !dataId.isEmpty() &&
               keyId != null && !keyId.isEmpty() &&
               ciphertext != null && !ciphertext.isEmpty() &&
               iv != null && !iv.isEmpty();
    }

    /**
     * 获取加密数据大小 (估算)
     */
    public int getEstimatedSize() {
        int size = 0;
        if (ciphertext != null) {
            size += ciphertext.length();
        }
        if (iv != null) {
            size += iv.length();
        }
        if (tag != null) {
            size += tag.length();
        }
        return size;
    }

    // === Getter/Setter ===

    public String getDataId() { return dataId; }
    public void setDataId(String dataId) { this.dataId = dataId; }

    public String getAlgorithm() { return algorithm; }
    public void setAlgorithm(String algorithm) { this.algorithm = algorithm; }

    public String getKeyId() { return keyId; }
    public void setKeyId(String keyId) { this.keyId = keyId; }

    public String getIv() { return iv; }
    public void setIv(String iv) { this.iv = iv; }

    public String getCiphertext() { return ciphertext; }
    public void setCiphertext(String ciphertext) { this.ciphertext = ciphertext; }

    public String getTag() { return tag; }
    public void setTag(String tag) { this.tag = tag; }

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }

    public String getDataType() { return dataType; }
    public void setDataType(String dataType) { this.dataType = dataType; }

    public String getChecksum() { return checksum; }
    public void setChecksum(String checksum) { this.checksum = checksum; }

    @NonNull
    @Override
    public String toString() {
        return "EncryptedData{" +
                "dataId='" + dataId + '\'' +
                ", algorithm='" + algorithm + '\'' +
                ", keyId='" + keyId + '\'' +
                ", size=" + getEstimatedSize() +
                '}';
    }
}