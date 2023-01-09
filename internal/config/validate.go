/*
 * auto - hypervisor agent for eve
 * Copyright (C) 2022-2023  BNS Services LLC

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package config

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Check for required fields in the config file
func validate() error {
	if err := validation.Validate(Config.Hostname, validation.Required, is.DNSName); err != nil {
		return fmt.Errorf("Configuration: hostname is not a hostname: %s", err)
	}

	if Config.TLSPath == "" {
		return fmt.Errorf("Configuration: TLSPath is required")
	}

	if err := validation.Validate(Config.API.Host, validation.Required, is.IP); err != nil {
		return fmt.Errorf("Configuration: API host is not an IP address: %s", err)
	}

	port := Config.API.Port
	if (port <= 1) || (port >= 65535) {
		return fmt.Errorf("Configuration: API port is not a valid port number: %d", port)
	}

	return nil
}
