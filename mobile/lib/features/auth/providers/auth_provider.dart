// ===== FILE: mobile/lib/features/auth/providers/auth_provider.dart =====
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/providers/providers.dart';
import '../../../core/constants/app_constants.dart';
import '../../../core/network/api_response.dart';
import '../../../data/models/user_model.dart';

class AuthState {
  final bool isAuthenticated;
  final bool isLoading;
  final UserModel? user;
  final String? errorMessage;

  AuthState({
    this.isAuthenticated = false,
    this.isLoading = false,
    this.user,
    this.errorMessage,
  });

  AuthState copyWith({
    bool? isAuthenticated,
    bool? isLoading,
    UserModel? user,
    String? errorMessage,
  }) {
    return AuthState(
      isAuthenticated: isAuthenticated ?? this.isAuthenticated,
      isLoading: isLoading ?? this.isLoading,
      user: user ?? this.user,
      errorMessage: errorMessage ?? this.errorMessage,
    );
  }
}

class AuthNotifier extends StateNotifier<AuthState> {
  final Ref ref;

  AuthNotifier(this.ref) : super(AuthState()) {
    _checkAuthStatus();
  }

  Future<void> _checkAuthStatus() async {
    final secureStorage = ref.read(secureStorageProvider);
    final token = await secureStorage.read(key: AppConstants.accessTokenKey);

    if (token != null) {
      try {
        final authDataSource = ref.read(authRemoteDataSourceProvider);
        final user = await authDataSource.getProfile();
        state = state.copyWith(
          isAuthenticated: true,
          user: user,
        );
      } catch (e) {
        await logout();
      }
    }
  }

  Future<void> login(String username, String password) async {
    state = state.copyWith(isLoading: true, errorMessage: null);

    try {
      final authDataSource = ref.read(authRemoteDataSourceProvider);
      final secureStorage = ref.read(secureStorageProvider);

      final loginResponse = await authDataSource.login(username, password);

      await secureStorage.write(
        key: AppConstants.accessTokenKey,
        value: loginResponse.accessToken,
      );
      await secureStorage.write(
        key: AppConstants.refreshTokenKey,
        value: loginResponse.refreshToken,
      );
      await secureStorage.write(
        key: AppConstants.userIdKey,
        value: loginResponse.user.id,
      );
      await secureStorage.write(
        key: AppConstants.userRoleKey,
        value: loginResponse.user.role,
      );
      await secureStorage.write(
        key: AppConstants.userNameKey,
        value: loginResponse.user.name,
      );

      state = state.copyWith(
        isAuthenticated: true,
        isLoading: false,
        user: loginResponse.user,
      );
    } on ApiException catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.message,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: 'Login gagal. Silakan coba lagi.',
      );
    }
  }

  Future<void> logout() async {
    state = state.copyWith(isLoading: true);

    try {
      final authDataSource = ref.read(authRemoteDataSourceProvider);
      await authDataSource.logout();
    } catch (e) {
      // Ignore error on logout
    }

    final secureStorage = ref.read(secureStorageProvider);
    await secureStorage.deleteAll();

    final localDb = ref.read(localDatabaseProvider);
    await localDb.clearAll();

    state = AuthState();
  }

  void clearError() {
    state = state.copyWith(errorMessage: null);
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref);
});
// ===== END FILE =====
