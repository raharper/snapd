// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package builtin_test

import (
	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/apparmor"
	"github.com/snapcore/snapd/interfaces/builtin"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/testutil"
)

type DesktopInterfaceSuite struct {
	iface    interfaces.Interface
	coreSlot *interfaces.Slot
	plug     *interfaces.Plug
}

var _ = Suite(&DesktopInterfaceSuite{
	iface: builtin.MustInterface("desktop"),
})

const desktopConsumerYaml = `name: consumer
apps:
 app:
  plugs: [desktop]
`

const desktopCoreYaml = `name: core
type: os
slots:
  desktop:
`

func (s *DesktopInterfaceSuite) SetUpTest(c *C) {
	s.plug = MockPlug(c, desktopConsumerYaml, nil, "desktop")
	s.coreSlot = MockSlot(c, desktopCoreYaml, nil, "desktop")
}

func (s *DesktopInterfaceSuite) TestName(c *C) {
	c.Assert(s.iface.Name(), Equals, "desktop")
}

func (s *DesktopInterfaceSuite) TestSanitizeSlot(c *C) {
	c.Assert(s.coreSlot.Sanitize(s.iface), IsNil)
	// desktop slot currently only used with core
	slot := &interfaces.Slot{SlotInfo: &snap.SlotInfo{
		Snap:      &snap.Info{SuggestedName: "some-snap"},
		Name:      "desktop",
		Interface: "desktop",
	}}
	c.Assert(slot.Sanitize(s.iface), ErrorMatches,
		"desktop slots are reserved for the core snap")
}

func (s *DesktopInterfaceSuite) TestSanitizePlug(c *C) {
	c.Assert(s.plug.Sanitize(s.iface), IsNil)
}

func (s *DesktopInterfaceSuite) TestAppArmorSpec(c *C) {
	// connected plug to core slot
	spec := &apparmor.Specification{}
	c.Assert(spec.AddConnectedPlug(s.iface, s.plug, nil, s.coreSlot, nil), IsNil)
	c.Assert(spec.SecurityTags(), DeepEquals, []string{"snap.consumer.app"})
	c.Assert(spec.SnippetForTag("snap.consumer.app"), testutil.Contains, "# Description: Can access basic graphical desktop resources")
	c.Assert(spec.SnippetForTag("snap.consumer.app"), testutil.Contains, "#include <abstractions/fonts>")
	c.Assert(spec.SnippetForTag("snap.consumer.app"), testutil.Contains, "/etc/gtk-3.0/settings.ini r,")

	// connected plug to core slot
	spec = &apparmor.Specification{}
	c.Assert(spec.AddConnectedSlot(s.iface, s.plug, nil, s.coreSlot, nil), IsNil)
	c.Assert(spec.SecurityTags(), HasLen, 0)
}

func (s *DesktopInterfaceSuite) TestStaticInfo(c *C) {
	si := interfaces.StaticInfoOf(s.iface)
	c.Assert(si.ImplicitOnCore, Equals, false)
	c.Assert(si.ImplicitOnClassic, Equals, true)
	c.Assert(si.Summary, Equals, `allows access to basic graphical desktop resources`)
	c.Assert(si.BaseDeclarationSlots, testutil.Contains, "desktop")
}

func (s *DesktopInterfaceSuite) TestInterfaces(c *C) {
	c.Check(builtin.Interfaces(), testutil.DeepContains, s.iface)
}
