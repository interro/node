/*
 * Copyright (C) 2017 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package dto

// SessionStats mapped to json structure
type SessionStats struct {
	BytesSent       uint64 `json:"bytes_sent"`
	BytesReceived   uint64 `json:"bytes_received"`
	ProviderID      string `json:"provider_id"`
	ConsumerCountry string `json:"consumer_country"`
	ServiceType     string `json:"service_type"`
}
