package layer

import (
	"archive/tar"
	"bytes"
)

func BaseLayerBytes() ([]byte, error) {
	bcdBytes, err := BaseLayerBCD()
	if err != nil {
		return nil, err
	}

	return windowsBaseLayerBytes(bcdBytes)
}

// Windows image layers must follow this pattern¹:
// - base layer² (always required; tar file with relative paths without "/" prefix; all parent directories require own tar entries)
//   \-> UtilityVM/Files/EFI/Microsoft/Boot/BCD   (file must exist and a valid BCD format - via `bcdedit` tool as below)
//   \-> Files/Windows/System32/config/DEFAULT   (file and must exist but can be empty)
//   \-> Files/Windows/System32/config/SAM       (file must exist but can be empty)
//   \-> Files/Windows/System32/config/SECURITY  (file must exist but can be empty)
//   \-> Files/Windows/System32/config/SOFTWARE  (file must exist but can be empty)
//   \-> Files/Windows/System32/config/SYSTEM    (file must exist but can be empty)
// - normal or top layer (optional; tar file with relative paths without "/" prefix; all parent directories require own tar entries)
//   \-> Files/                   (required directory entry)
//   \-> Files/mystuff.exe        (optional container filesystem files - C:\mystuff.exe)
//   \-> Hives/                   (required directory entry)
//   \-> Hives/DefaultUser_Delta  (optional Windows reg hive delta; BCD format - HKEY_USERS\.DEFAULT additional content)
//   \-> Hives/Sam_Delta          (optional Windows reg hive delta; BCD format - HKEY_LOCAL_MACHINE\SAM additional content)
//   \-> Hives/Security_Delta     (optional Windows reg hive delta; BCD format - HKEY_LOCAL_MACHINE\SECURITY additional content)
//   \-> Hives/Software_Delta     (optional Windows reg hive delta; BCD format - HKEY_LOCAL_MACHINE\SOFTWARE additional content)
//   \-> Hives/System_Delta       (optional Windows reg hive delta; BCD format - HKEY_LOCAL_MACHINE\SYSTEM additional content)
// 1. This was all discovered experimentally and should be considered an undocumented API, subject to change when the Windows Daemon internals change
// 2. There are many other files in an "real" base layer but this is the minimum set which a Daemon can store and use to create an container
func windowsBaseLayerBytes(bcdBytes []byte) ([]byte, error) {
	layerBuffer := &bytes.Buffer{}
	tw := tar.NewWriter(layerBuffer)

	_ = bcdBytes
	if err := tw.WriteHeader(&tar.Header{Name: "UtilityVM", Typeflag: tar.TypeDir}); err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "UtilityVM/Files", Typeflag: tar.TypeDir}); err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "UtilityVM/Files/EFI", Typeflag: tar.TypeDir}); err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "UtilityVM/Files/EFI/Microsoft", Typeflag: tar.TypeDir}); err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "UtilityVM/Files/EFI/Microsoft/Boot", Typeflag: tar.TypeDir}); err != nil {
		return nil, err
	}

	if err := tw.WriteHeader(&tar.Header{Name: "UtilityVM/Files/EFI/Microsoft/Boot/BCD", Size: int64(len(bcdBytes)), Mode: 0644}); err != nil {
		return nil, err
	}
	if _, err := tw.Write(bcdBytes); err != nil {
		return nil, err
	}

	if err := tw.WriteHeader(&tar.Header{Name: "Files", Typeflag: tar.TypeDir}); err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "Files/Windows", Typeflag: tar.TypeDir}); err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "Files/Windows/System32", Typeflag: tar.TypeDir}); err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "Files/Windows/System32/config", Typeflag: tar.TypeDir}); err != nil {
		return nil, err
	}

	if err := tw.WriteHeader(&tar.Header{Name: "Files/Windows/System32/config/DEFAULT", Size: 0, Mode: 0644}); err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "Files/Windows/System32/config/SAM", Size: 0, Mode: 0644}); err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "Files/Windows/System32/config/SECURITY", Size: 0, Mode: 0644}); err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "Files/Windows/System32/config/SOFTWARE", Size: 0, Mode: 0644}); err != nil {
		return nil, err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "Files/Windows/System32/config/SYSTEM", Size: 0, Mode: 0644}); err != nil {
		return nil, err
	}

	return layerBuffer.Bytes(), nil
}
