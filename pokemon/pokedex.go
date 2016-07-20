// showdown2irc - use Showdown chat with an IRC client
// Copyright (C) 2016 Konrad Borowski
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package pokemon

import "github.com/xfix/showdown2irc/showdown"

//go:generate node create_definitions

type Pokemon struct {
	Species   string
	Types     []Type
	Abilities []string
	BaseStats Stats
}

type Stats struct {
	Hp, Atk, Def, Spa, Spd, Spe int
}

type Type int

const (
	Normal Type = iota
	Fighting
	Flying
	Poison
	Ground
	Rock
	Bug
	Ghost
	Steel
	Fire
	Water
	Grass
	Electric
	Psychic
	Ice
	Dragon
	Dark
	Fairy
	Bird // MissingNo.'s type
)

const Hidden = 2

type Move struct {
	Name           string
	Type           Type
	DamageCategory DamageCategory
	BasePower      int
	Accuracy       int
	PP             int
	Description    string
}

type DamageCategory int

const (
	Physical DamageCategory = iota
	Special
	Status
)

func GetPokemon(name showdown.UserID) *Pokemon {
	return pokemon[name]
}

func GetMove(move showdown.UserID) *Move {
	return moves[move]
}

func GetAbilityDescription(ability showdown.UserID) string {
	return abilityDescriptions[ability]
}
