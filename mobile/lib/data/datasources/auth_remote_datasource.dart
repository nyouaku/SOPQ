// ===== FILE: mobile/lib/data/datasources/auth_remote_datasource.dart =====
import 'package:dio/dio.dart';
import '../../core/network/api_response.dart';
import '../../core/constants/app_constants.dart';
import '../models/user_model.dart';

class AuthRemoteDataSource {
  final Dio _dio;

  AuthRemoteDataSource(this._dio);

  Future<LoginResponse> login(String username, String password) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.login,
        data: {
          'username': username,
          'password': password,
        },
      );

      final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
        response.data,
        (data) => data as Map<String, dynamic>,
      );

      if (apiResponse.success && apiResponse.data != null) {
        return LoginResponse.fromJson(apiResponse.data!);
      } else {
        throw ApiException(
          message: apiResponse.error?.message ?? 'Login failed',
          code: apiResponse.error?.code,
        );
      }
    } on DioException catch (e) {
      throw ApiException(
        message: e.response?.data['error']?['message'] ?? e.message ?? 'Network error',
        statusCode: e.response?.statusCode,
      );
    }
  }

  Future<void> logout() async {
    try {
      await _dio.post(ApiEndpoints.logout);
    } on DioException catch (e) {
      throw ApiException(
        message: e.response?.data['error']?['message'] ?? e.message ?? 'Logout failed',
        statusCode: e.response?.statusCode,
      );
    }
  }

  Future<UserModel> getProfile() async {
    try {
      final response = await _dio.get(ApiEndpoints.profile);

      final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
        response.data,
        (data) => data as Map<String, dynamic>,
      );

      if (apiResponse.success && apiResponse.data != null) {
        return UserModel.fromJson(apiResponse.data!);
      } else {
        throw ApiException(
          message: apiResponse.error?.message ?? 'Failed to get profile',
          code: apiResponse.error?.code,
        );
      }
    } on DioException catch (e) {
      throw ApiException(
        message: e.response?.data['error']?['message'] ?? e.message ?? 'Network error',
        statusCode: e.response?.statusCode,
      );
    }
  }
}
// ===== END FILE =====
