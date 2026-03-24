cmake_minimum_required(VERSION 3.1)
project(prometheus_monitor)

set(CMAKE_CXX_STANDARD 14)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(PkgConfig REQUIRED)

include(FetchContent)
FetchContent_Declare(
    prometheus-cpp
    GIT_REPOSITORY https://github.com/jupp0r/prometheus-cpp.git
    GIT_TAG        v1.3.0
)

set(ENABLE_PUSH ON)
set(ENABLE_COMPRESSION OFF)
set(ENABLE_TESTING OFF)
FetchContent_MakeAvailable(prometheus-cpp)

add_executable(main main.cpp)
target_link_libraries(main PRIVATE
    prometheus-cpp::core
    prometheus-cpp::pull
    prometheus-cpp::push
)
