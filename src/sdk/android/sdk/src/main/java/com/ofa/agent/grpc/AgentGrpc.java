package com.ofa.agent.grpc;

import io.grpc.stub.AbstractStub;
import io.grpc.Channel;
import io.grpc.CallOptions;
import io.grpc.stub.StreamObserver;

/**
 * Stub class for Agent gRPC service.
 * This is a placeholder - actual implementation requires protobuf generation.
 */
public class AgentGrpc {

    public static final String SERVICE_NAME = "Agent";

    public static AgentStub newStub(Channel channel) {
        return new AgentStub(channel);
    }

    public static class AgentStub extends AbstractStub<AgentStub> {

        public AgentStub(Channel channel) {
            super(channel);
        }

        public AgentStub(Channel channel, CallOptions callOptions) {
            super(channel, callOptions);
        }

        @Override
        protected AgentStub build(Channel channel, CallOptions callOptions) {
            return new AgentStub(channel, callOptions);
        }

        public StreamObserver<AgentOuterClass.AgentMessage> connect(
                StreamObserver<AgentOuterClass.CenterMessage> responseObserver) {
            // Stub implementation
            return new StreamObserver<AgentOuterClass.AgentMessage>() {
                @Override
                public void onNext(AgentOuterClass.AgentMessage value) {}

                @Override
                public void onError(Throwable t) {}

                @Override
                public void onCompleted() {}
            };
        }
    }
}