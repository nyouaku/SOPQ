<script lang="ts">
	import { onMount } from 'svelte';
	import { themeStore } from '$lib/stores/theme';
	import type { Theme } from '$lib/stores/theme';
	import { authStore } from '$lib/stores/auth';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import ThemeSelector from '$lib/components/ThemeSelector.svelte';
	import LoaderSettings from '$lib/components/LoaderSettings.svelte';
	import SoundSettingsPanel from '$lib/components/SoundSettingsPanel.svelte';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';

	// Auth guard terpusat di src/routes/+layout.svelte.
	let currentTheme: Theme = { layout: 'bottom', color: 'rack-room' };
	themeStore.subscribe((t) => (currentTheme = t));

	// Reaktif: user baru terisi setelah checkAuth selesai (penting di mobile).
	$: isAdmin = $authStore.user?.role?.name === 'admin';

	// --- Manajemen staff (admin) ---
	let staff: any[] = [];
	let roles: any[] = [];
	let isLoadingStaff = false;
	let staffError = '';

	let showAddStaff = false;
	let isSavingStaff = false;
	let staffForm = { username: '', full_name: '', password: '', role_id: 0 };
	let staffFeedback = '';

	onMount(async () => {
		if (!isAdmin) return;
		await Promise.all([loadStaff(), loadRoles()]);
	});

	async function loadStaff() {
		isLoadingStaff = true;
		staffError = '';
		try {
			const res = await apiClient.get<{ users: any[] }>(API_ENDPOINTS.users);
			if (res.success && res.data) {
				// Sembunyikan diri sendiri dari daftar.
				staff = res.data.users.filter((u) => u.id !== $authStore.user?.id);
			}
		} catch (e: any) {
			staffError = e.message || 'Gagal memuat staff';
		} finally {
			isLoadingStaff = false;
		}
	}

	async function loadRoles() {
		try {
			const res = await apiClient.get<{ roles: any[] }>(API_ENDPOINTS.roles);
			if (res.success && res.data) {
				roles = res.data.roles;
				if (roles.length > 0 && !staffForm.role_id) {
					staffForm.role_id = roles[0].id;
				}
			}
		} catch (e: any) {
			console.error('Failed to load roles:', e);
		}
	}

	function openAddStaff() {
		staffForm = { username: '', full_name: '', password: '', role_id: roles[0]?.id || 0 };
		staffFeedback = '';
		showAddStaff = true;
	}

	async function handleAddStaff() {
		if (!staffForm.username || !staffForm.full_name || !staffForm.password || !staffForm.role_id) {
			staffFeedback = 'Semua kolom wajib diisi';
			return;
		}
		isSavingStaff = true;
		staffFeedback = '';
		try {
			const res = await apiClient.post(API_ENDPOINTS.register, {
				username: staffForm.username,
				full_name: staffForm.full_name,
				password: staffForm.password,
				role_id: staffForm.role_id
			});
			if (res.success) {
				staffFeedback = 'Staff berhasil ditambahkan';
				showAddStaff = false;
				await loadStaff();
			} else {
				staffFeedback = res.error?.message || 'Gagal menambahkan staff';
			}
		} catch (e: any) {
			staffFeedback = e.message || 'Gagal menambahkan staff';
		} finally {
			isSavingStaff = false;
		}
	}

	async function toggleUserStatus(user: any) {
		const action = user.is_active ? 'menonaktifkan' : 'mengaktifkan';
		if (!confirm(`Yakin ingin ${action} user ${user.full_name || user.username}?`)) return;
		try {
			const res = await apiClient.patch(`${API_ENDPOINTS.users}/${user.id}/toggle-status`, {
				is_active: !user.is_active
			});
			if (res.success) {
				toastStore.success(`User berhasil ${user.is_active ? 'dinonaktifkan' : 'diaktifkan'}`);
				await loadStaff();
			} else {
				toastStore.error(res.error?.message || 'Gagal mengubah status');
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal mengubah status');
		}
	}

	async function resetPassword(user: any) {
		if (!confirm(`Yakin ingin reset password ${user.full_name || user.username}? Semua sesi aktif akan dicabut.`)) return;
		try {
			const res = await apiClient.post(`${API_ENDPOINTS.users}/${user.id}/reset-password`, {});
			if (res.success && res.data?.new_password) {
				// Tampilkan password baru
				const newPassword = res.data.new_password;
				showResetPasswordModal(newPassword);
				await loadStaff();
			} else {
				toastStore.error(res.error?.message || 'Gagal reset password');
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal reset password');
		}
	}

	let newPasswordDisplay = '';
	let showNewPassword = false;

	function showResetPasswordModal(password: string) {
		newPasswordDisplay = password;
		showNewPassword = true;
	}
</script>

<svelte:head>
	<title>Settings - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="mb-6">
			<h1 class="page-title">Settings</h1>
			<p class="mt-1 text-sm text-base-content/60">Tampilan, akun, dan staff</p>
		</div>

		<div class="space-y-6">
			<!-- Tampilan -->
			<section class="rounded-2xl border border-base-300 bg-base-200">
				<header class="border-b border-base-300 px-4 py-3 md:px-5">
					<h2 class="font-semibold">Tampilan</h2>
					<p class="text-xs text-base-content/60">Perubahan langsung diterapkan</p>
				</header>
				<div class="p-4 md:p-5">
					<ThemeSelector currentLayout={currentTheme.layout} currentColor={currentTheme.color} />
				</div>
			</section>

			<!-- Preloader & Loader -->
			<section class="rounded-2xl border border-base-300 bg-base-200">
				<header class="border-b border-base-300 px-4 py-3 md:px-5">
					<h2 class="font-semibold">Preloader & Loader</h2>
					<p class="text-xs text-base-content/60">Pilih gaya animasi memuat</p>
				</header>
				<div class="p-4 md:p-5">
					<LoaderSettings />
				</div>
			</section>

			<!-- Suara & Notifikasi -->
			<section class="rounded-2xl border border-base-300 bg-base-200">
				<header class="border-b border-base-300 px-4 py-3 md:px-5">
					<h2 class="font-semibold">Suara & Notifikasi</h2>
					<p class="text-xs text-base-content/60">Atur efek suara scan, simpan, error</p>
				</header>
				<div class="p-4 md:p-5">
					<SoundSettingsPanel />
				</div>
			</section>

			<!-- Staff (admin only) -->
			{#if isAdmin}
				<section class="rounded-2xl border border-base-300 bg-base-200">
					<header class="flex items-center justify-between gap-3 border-b border-base-300 px-4 py-3 md:px-5">
						<div>
							<h2 class="font-semibold">Staff</h2>
							<p class="text-xs text-base-content/60">Akun pengguna: admin, supervisor, staff gudang</p>
						</div>
						<button class="btn btn-primary btn-sm" on:click={openAddStaff}>
							<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
							</svg>
							Tambah staff
						</button>
					</header>

					<div class="p-4 md:p-5">
						{#if isLoadingStaff}
							<div class="flex justify-center py-8">
								<Loader size={32} />
							</div>
						{:else if staffError}
							<p class="text-sm text-error">{staffError}</p>
						{:else if staff.length === 0}
							<p class="text-sm text-base-content/60">Belum ada staff lain.</p>
						{:else}
							<ul class="divide-y divide-base-300">
								{#each staff as user}
									<li class="flex items-center gap-3 py-3">
										<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-base-300 text-sm font-semibold uppercase">
											{(user.full_name || user.username).charAt(0)}
										</div>
										<div class="min-w-0 flex-1">
											<div class="truncate text-sm font-medium">
												{user.full_name || user.username}
											</div>
											<div class="metric truncate text-xs text-base-content/60">@{user.username}</div>
										</div>
										<span class="badge badge-ghost badge-sm capitalize">
											{user.role?.name || '-'}
										</span>
										<!-- Status badge -->
										<span class="badge badge-sm {user.is_active ? 'badge-success' : 'badge-error'}">
											{user.is_active ? 'Aktif' : 'Nonaktif'}
										</span>
										<!-- Actions -->
										<div class="flex shrink-0 gap-1">
											<button
												class="btn btn-ghost btn-xs btn-square"
												title={user.is_active ? 'Nonaktifkan' : 'Aktifkan'}
												aria-label={user.is_active ? `Nonaktifkan ${user.full_name || user.username}` : `Aktifkan ${user.full_name || user.username}`}
												on:click={() => toggleUserStatus(user)}
											>
												{#if user.is_active}
													<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
														<path stroke-linecap="round" stroke-linejoin="round" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728L5.636 5.636" />
													</svg>
												{:else}
													<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
														<path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
													</svg>
												{/if}
											</button>
											<button
												class="btn btn-ghost btn-xs btn-square"
												title="Reset password"
												aria-label={`Reset password ${user.full_name || user.username}`}
												on:click={() => resetPassword(user)}
											>
												<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
													<path stroke-linecap="round" stroke-linejoin="round" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
												</svg>
											</button>
										</div>
									</li>
								{/each}
							</ul>
						{/if}
					</div>
				</section>
			{/if}

			<!-- Akun -->
			{#if $authStore.user}
				<section class="rounded-2xl border border-base-300 bg-base-200">
					<header class="border-b border-base-300 px-4 py-3 md:px-5">
						<h2 class="font-semibold">Akun</h2>
						<p class="text-xs text-base-content/60">Data pengguna yang sedang masuk</p>
					</header>

					<div class="p-4 md:p-5">
						<div class="mb-5 flex items-center gap-3">
							<div class="avatar placeholder">
								<div class="w-12 rounded-full bg-primary text-primary-content">
									<span class="text-lg">
										{($authStore.user.full_name || $authStore.user.username)?.[0]?.toUpperCase()}
									</span>
								</div>
							</div>
							<div class="min-w-0">
								<div class="truncate font-semibold">
									{$authStore.user.full_name || $authStore.user.username}
								</div>
								<div class="text-xs capitalize text-base-content/60">
									{$authStore.user.role?.name || '-'}
								</div>
							</div>
						</div>

						<dl class="grid grid-cols-1 gap-x-6 gap-y-4 text-sm sm:grid-cols-2">
							<div>
								<dt class="text-xs text-base-content/60">Username</dt>
								<dd class="metric mt-0.5">{$authStore.user.username}</dd>
							</div>
							<div>
								<dt class="text-xs text-base-content/60">Nama lengkap</dt>
								<dd class="mt-0.5 truncate">{$authStore.user.full_name || '-'}</dd>
							</div>
							<div>
								<dt class="text-xs text-base-content/60">Role</dt>
								<dd class="mt-0.5 capitalize">{$authStore.user.role?.name || '-'}</dd>
							</div>
							<div>
								<dt class="text-xs text-base-content/60">Email</dt>
								<dd class="mt-0.5 truncate">{$authStore.user.email || '-'}</dd>
							</div>
						</dl>

						<div class="mt-5 border-t border-base-300 pt-4">
							<button class="btn btn-outline btn-error btn-sm" on:click={() => authStore.logout()}>
								Logout
							</button>
						</div>
					</div>
				</section>
			{/if}
		</div>
	</div>
</LayoutWrapper>

<!-- Modal: tambah staff -->
{#if showAddStaff}
	<div class="modal modal-open" role="dialog" aria-modal="true" aria-label="Tambah staff">
		<div class="modal-box max-w-md">
			<h3 class="text-lg font-semibold">Tambah Staff</h3>
			<p class="mt-1 text-sm text-base-content/60">Buat akun baru untuk admin, supervisor, atau staff gudang.</p>

			<form class="mt-4 space-y-3" on:submit|preventDefault={handleAddStaff}>
				<div class="form-control">
					<label class="label py-0" for="staff-username">
						<span class="label-text text-xs text-base-content/60">Username</span>
					</label>
					<input
						id="staff-username"
						type="text"
						class="input input-bordered input-sm"
						placeholder="mis. budi"
						bind:value={staffForm.username}
						required
					/>
				</div>

				<div class="form-control">
					<label class="label py-0" for="staff-fullname">
						<span class="label-text text-xs text-base-content/60">Nama lengkap</span>
					</label>
					<input
						id="staff-fullname"
						type="text"
						class="input input-bordered input-sm"
						placeholder="mis. Budi Santoso"
						bind:value={staffForm.full_name}
						required
					/>
				</div>

				<div class="form-control">
					<label class="label py-0" for="staff-password">
						<span class="label-text text-xs text-base-content/60">Password</span>
					</label>
					<input
						id="staff-password"
						type="password"
						autocomplete="new-password"
						class="input input-bordered input-sm"
						placeholder="min. 6 karakter"
						bind:value={staffForm.password}
						required
					/>
				</div>

				<div class="form-control">
					<label class="label py-0" for="staff-role">
						<span class="label-text text-xs text-base-content/60">Role</span>
					</label>
					<select
						id="staff-role"
						class="select select-bordered select-sm w-full"
						bind:value={staffForm.role_id}
					>
						{#each roles as role}
							<option value={role.id}>{role.name}</option>
						{/each}
					</select>
				</div>

				{#if staffFeedback}
					<div class="alert alert-error py-2 text-sm">
						<span>{staffFeedback}</span>
					</div>
				{/if}

				<div class="modal-action">
					<button type="button" class="btn" on:click={() => (showAddStaff = false)}>Batal</button>
					<button type="submit" class="btn btn-primary" disabled={isSavingStaff}>
						{#if isSavingStaff}
							<Loader size={20} />
						{/if}
						Simpan
					</button>
				</div>
			</form>
		</div>
	</div>
	<button class="modal-backdrop" aria-label="Tutup tambah staff" on:click={() => (showAddStaff = false)}></button>
{/if}

<!-- Modal tampilkan password baru -->
{#if showNewPassword}
	<div class="modal modal-open">
		<div class="modal-box max-w-md">
			<h3 class="text-lg font-bold">Password Baru</h3>
			<p class="mt-2 text-sm text-base-content/70">
				Simpan password ini. Sesi lama user sudah dicabut, user harus login ulang dengan password ini.
			</p>
			<div class="mt-4 rounded-xl border-2 border-dashed border-primary bg-primary/5 p-4 text-center">
				<p class="metric text-2xl font-bold tracking-wider select-all">{newPasswordDisplay}</p>
			</div>
			<div class="modal-action">
				<button
					class="btn btn-primary"
					on:click={() => {
						navigator.clipboard.writeText(newPasswordDisplay);
						toastStore.success('Password disalin ke clipboard');
					}}
				>
					Salin
				</button>
				<button class="btn btn-ghost" on:click={() => (showNewPassword = false)}>Tutup</button>
			</div>
		</div>
	</div>
	<button class="modal-backdrop" aria-label="Tutup" on:click={() => (showNewPassword = false)}></button>
{/if}
