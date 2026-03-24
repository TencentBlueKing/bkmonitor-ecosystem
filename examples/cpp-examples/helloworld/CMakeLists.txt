cmake_minimum_required(VERSION 3.1)
project(HelloWorld)

set(CMAKE_CXX_STANDARD 17)
set(project_name helloworld)

include_directories(include)

set(SOURCES
        src/main.cpp
        src/server.cpp
        src/client.cpp
)
add_executable(helloworld ${SOURCES})

set(OATPP_ROOT ${CMAKE_SOURCE_DIR}/oatpp)
find_library(OATPP_LIB NAMES liboatpp.a HINTS "${OATPP_ROOT}/build/src/" NO_DEFAULT_PATH)
if (NOT OATPP_LIB)
    message(SEND_ERROR "Did not find oatpp library ${OATPP_ROOT}/build/src")
endif()
# set the path to the directory containing "oatpp" package configuration files
include_directories(${OATPP_ROOT}/src)

find_package(nlohmann_json REQUIRED)
find_package(CURL REQUIRED)
find_package(gRPC REQUIRED)
find_package(Protobuf REQUIRED)

set(opentelemetry-cpp_DIR ${CMAKE_SOURCE_DIR}/otel-cpp/lib/cmake/opentelemetry-cpp)
find_package(opentelemetry-cpp CONFIG REQUIRED NO_DEFAULT_PATH)

target_link_libraries(helloworld
        protobuf::libprotobuf
        CURL::libcurl
        ${OATPP_LIB}
        ${OPENTELEMETRY_CPP_LIBRARIES}
        gRPC::grpc++
        nlohmann_json::nlohmann_json
)
install(TARGETS helloworld DESTINATION bin)
