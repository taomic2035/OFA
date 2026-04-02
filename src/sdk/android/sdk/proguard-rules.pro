# OFA Agent SDK ProGuard Rules

# Keep LLM classes
-keep class com.ofa.agent.llm.** { *; }

# Keep MCP classes
-keep class com.ofa.agent.mcp.** { *; }

# Keep tool classes
-keep class com.ofa.agent.tool.** { *; }

# Keep gRPC generated classes
-keep class com.ofa.agent.grpc.** { *; }

# Keep protobuf classes
-keep class * extends com.google.protobuf.GeneratedMessageLite { *; }

# Keep TensorFlow Lite
-keep class org.tensorflow.lite.** { *; }

# Gson
-keepattributes Signature
-keepattributes *Annotation*
-keep class com.google.gson.** { *; }
-keep class * implements com.google.gson.TypeAdapterFactory
-keep class * implements com.google.gson.JsonSerializer
-keep class * implements com.google.gson.JsonDeserializer