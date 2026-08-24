cmake_minimum_required(VERSION 3.10)
project(log_reporter LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 指定输出目录
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)

find_package(CURL REQUIRED)
find_path(NLOHMANN_JSON_INCLUDE_DIR nlohmann/json.hpp
    PATHS /usr/include /usr/local/include)

add_executable(log_reporter main.cpp)

target_link_libraries(log_reporter PRIVATE ${CURL_LIBRARIES})
target_include_directories(log_reporter PRIVATE
    ${CURL_INCLUDE_DIRS}
    ${NLOHMANN_JSON_INCLUDE_DIR}
)
