# OFA C++ Agent SDK

cmake_minimum_required(VERSION 3.14)
project(ofa_agent VERSION 8.1.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Options
option(OFA_BUILD_EXAMPLES "Build example programs" ON)
option(OFA_BUILD_TESTS "Build tests" OFF)
option(OFA_ENABLE_GRPC "Enable gRPC support" ON)
option(OFA_ENABLE_MQTT "Enable MQTT support" OFF)

# Dependencies
find_package(Threads REQUIRED)

if(OFA_ENABLE_GRPC)
    find_package(gRPC CONFIG REQUIRED)
    find_package(Protobuf CONFIG REQUIRED)
endif()

# Library sources
set(OFA_SOURCES
    src/agent.cpp
    src/config.cpp
    src/skills.cpp
    src/connection.cpp
    src/message.cpp
    src/protocol.cpp
    src/builtin.cpp
)

# Library headers
set(OFA_HEADERS
    include/ofa/agent.hpp
    include/ofa/config.hpp
    include/ofa/skills.hpp
    include/ofa/connection.hpp
    include/ofa/message.hpp
    include/ofa/protocol.hpp
    include/ofa/types.hpp
    include/ofa/error.hpp
)

# Create library
add_library(ofa_agent STATIC ${OFA_SOURCES} ${OFA_HEADERS})

target_include_directories(ofa_agent
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(ofa_agent
    PUBLIC
        Threads::Threads
        nlohmann_json::nlohmann_json
)

if(OFA_ENABLE_GRPC)
    target_link_libraries(ofa_agent
        PUBLIC
            gRPC::grpc++
            gRPC::grpc++_reflection
    )
    target_compile_definitions(ofa_agent PUBLIC OFA_ENABLE_GRPC=1)
endif()

# Fetch nlohmann_json
include(FetchContent)
FetchContent_Declare(
    json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.2
)
FetchContent_MakeAvailable(json)

# Examples
if(OFA_BUILD_EXAMPLES)
    add_executable(example_agent examples/example.cpp)
    target_link_libraries(example_agent PRIVATE ofa_agent)
endif()

# Tests
if(OFA_BUILD_TESTS)
    enable_testing()
    add_subdirectory(tests)
endif()

# Install
install(TARGETS ofa_agent
    EXPORT ofa_agent-targets
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
    INCLUDES DESTINATION include
)

install(DIRECTORY include/ofa DESTINATION include)

# Package configuration
include(CMakePackageConfigHelpers)
configure_package_config_file(
    ${CMAKE_CURRENT_SOURCE_DIR}/cmake/ofa_agent-config.cmake.in
    ${CMAKE_CURRENT_BINARY_DIR}/ofa_agent-config.cmake
    INSTALL_DESTINATION lib/cmake/ofa_agent
)

install(FILES
    ${CMAKE_CURRENT_BINARY_DIR}/ofa_agent-config.cmake
    DESTINATION lib/cmake/ofa_agent
)