--  This file is part of the eliona project.
--  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
--  ______ _ _
-- |  ____| (_)
-- | |__  | |_  ___  _ __   __ _
-- |  __| | | |/ _ \| '_ \ / _` |
-- | |____| | | (_) | | | | (_| |
-- |______|_|_|\___/|_| |_|\__,_|
--
--  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
--  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
--  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
--  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
--  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

create schema if not exists kentix_one;

-- Configuration corresponds to one Kentix device
-- Should be editable by eliona frontend.
create table if not exists kentix_one.configuration
(
	id               bigserial primary key,
	address          text,
	api_key          text,
	enable           boolean default false,
	refresh_interval integer not null default 60,
	request_timeout  integer not null default 120,
	active           boolean default false,
	project_ids      text[]
);

-- Device corresponds to one asset in Eliona
-- Should be read-only by eliona frontend.
create table if not exists kentix_one.device
(
	configuration_id bigserial references kentix_one.configuration(id),
	project_id       text      not null,
	serial_number    text      not null,
	asset_id         integer,
	primary key (configuration_id, project_id, serial_number)
);

-- Makes the new objects available for all other init steps
commit;
