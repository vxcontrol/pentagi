//go:build windows
// +build windows

package system

import (
	"fmt"

	"github.com/go-ole/go-ole"
	"golang.org/x/sys/windows/registry"
)

func getMachineID() (string, error) {
	sp, _ := getSystemProduct()
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Cryptography`,
		registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return sp, err
	}
	defer k.Close()

	s, _, err := k.GetStringValue("MachineGuid")
	if err != nil {
		return sp, err
	}
	return s + ":" + sp, nil
}

func getSystemProduct() (string, error) {
	var err error
	var classID *ole.GUID

	IID_ISWbemLocator, _ := ole.CLSIDFromString("{76A6415B-CB41-11D1-8B02-00600806D9B6}")

	err = ole.CoInitialize(0)
	if err != nil {
		return "", fmt.Errorf("OLE initialize error: %v", err)
	}
	defer ole.CoUninitialize()

	classID, err = ole.ClassIDFrom("WbemScripting.SWbemLocator")
	if err != nil {
		return "", fmt.Errorf("CreateObject WbemScripting.SWbemLocator returned with %v", err)
	}

	comserver, err := ole.CreateInstance(classID, ole.IID_IUnknown)
	if err != nil {
		return "", fmt.Errorf("CreateInstance WbemScripting.SWbemLocator returned with %v", err)
	}
	if comserver == nil {
		return "", fmt.Errorf("CreateObject WbemScripting.SWbemLocator not an object")
	}
	defer comserver.Release()

	dispatch, err := comserver.QueryInterface(IID_ISWbemLocator)
	if err != nil {
		return "", fmt.Errorf("context.iunknown.QueryInterface returned with %v", err)
	}
	defer dispatch.Release()

	wbemServices, err := dispatch.CallMethod("ConnectServer")
	if err != nil {
		return "", fmt.Errorf("ConnectServer failed with %v", err)
	}
	defer wbemServices.Clear()

	query := "SELECT * FROM Win32_ComputerSystemProduct"
	objectset, err := wbemServices.ToIDispatch().CallMethod("ExecQuery", query)
	if err != nil {
		return "", fmt.Errorf("ExecQuery failed with %v", err)
	}
	defer objectset.Clear()

	enum_property, err := objectset.ToIDispatch().GetProperty("_NewEnum")
	if err != nil {
		return "", fmt.Errorf("Get _NewEnum property failed with %v", err)
	}
	defer enum_property.Clear()

	enum, err := enum_property.ToIUnknown().IEnumVARIANT(ole.IID_IEnumVariant)
	if err != nil {
		return "", fmt.Errorf("IEnumVARIANT() returned with %v", err)
	}
	if enum == nil {
		return "", fmt.Errorf("Enum is nil")
	}
	defer enum.Release()

	for tmp, length, err := enum.Next(1); length > 0; tmp, length, err = enum.Next(1) {
		if err != nil {
			return "", fmt.Errorf("Next() returned with %v", err)
		}
		tmp_dispatch := tmp.ToIDispatch()
		defer tmp_dispatch.Release()

		props, err := tmp_dispatch.GetProperty("Properties_")
		if err != nil {
			return "", fmt.Errorf("Get Properties_ property failed with %v", err)
		}
		defer props.Clear()

		props_enum_property, err := props.ToIDispatch().GetProperty("_NewEnum")
		if err != nil {
			return "", fmt.Errorf("Get _NewEnum property failed with %v", err)
		}
		defer props_enum_property.Clear()

		props_enum, err := props_enum_property.ToIUnknown().IEnumVARIANT(ole.IID_IEnumVariant)
		if err != nil {
			return "", fmt.Errorf("IEnumVARIANT failed with %v", err)
		}
		defer props_enum.Release()

		class_variant, err := tmp_dispatch.GetProperty("UUID")
		if err != nil {
			return "", fmt.Errorf("Get UUID property failed with %v", err)
		}
		defer class_variant.Clear()

		class_name := class_variant.ToString()
		return class_name, nil
	}

	return "", fmt.Errorf("not found")
}
