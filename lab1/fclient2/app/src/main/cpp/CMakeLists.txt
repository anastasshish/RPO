cmake_minimum_required(VERSION 3.22.1)

project("fclient2")


add_library(fclient2 SHARED
    native-lib.cpp
)


add_library(spdlog STATIC IMPORTED)
set_target_properties(spdlog PROPERTIES IMPORTED_LOCATION
    "${CMAKE_SOURCE_DIR}/../../../../../libs/spdlog/build/${ANDROID_ABI}/libspdlog.a"
)


add_library(mbedcrypto SHARED IMPORTED)
set_target_properties(mbedcrypto PROPERTIES IMPORTED_LOCATION
    ${CMAKE_CURRENT_SOURCE_DIR}/../nativeLibs/${ANDROID_ABI}/libmbedcrypto.so
)


target_include_directories(fclient2 PUBLIC
    "C:/Users/shish/AndroidStudioProjects/libs/spdlog/include"
    "C:/Users/shish/AndroidStudioProjects/libs/mbedtls/include"
)


find_library(log-lib log)
find_library(android-lib android)


target_link_libraries(fclient2
    spdlog
    mbedcrypto
    ${android-lib}
    ${log-lib}
)